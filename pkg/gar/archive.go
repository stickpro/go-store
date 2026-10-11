// Package gar reads the ФНС «Государственный адресный реестр» (ГАР БД ФИАС)
// XML export, gar_xml.zip, published at https://fias.nalog.ru.
//
// The full archive is ~58 GB, but a settlement import only needs a few
// per-region files (AS_ADDR_OBJ, AS_ADDR_OBJ_PARAMS, AS_ADM_HIERARCHY, ~3 GB
// compressed). An Archive therefore opens either a local zip or the zip's
// URL directly: for a URL only the central directory and the requested
// entries are fetched with HTTP range requests.
package gar

import (
	"archive/zip"
	"compress/flate"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

// Per-region file kinds, as named in the archive (NN/<KIND>_YYYYMMDD_<guid>.XML).
const (
	KindAddrObj       = "AS_ADDR_OBJ"
	KindAddrObjParams = "AS_ADDR_OBJ_PARAMS"
	KindAdmHierarchy  = "AS_ADM_HIERARCHY"
)

// Object levels (AS_OBJECT_LEVELS).
const (
	LevelRegion     = 1 // Субъект РФ
	LevelAdmArea    = 2 // Административный район
	LevelMunArea    = 3 // Муниципальный район
	LevelCity       = 5 // Город
	LevelSettlement = 6 // Населенный пункт
)

// Parameter types (AS_PARAM_TYPES).
const (
	ParamIFNSFL       = 1  // ИФНС ФЛ
	ParamPostIndex    = 5  // Почтовый индекс
	ParamOKATO        = 6  // ОКАТО
	ParamOKTMO        = 7  // ОКТМО
	ParamKLADR        = 10 // Код КЛАДР
	ParamIsAdmCapital = 22 // Административный центр субъекта РФ
	ParamIsMunCapital = 23 // Административный центр муниципального образования
)

// LatestInfoURL returns metadata of the most recent ФНС export.
const LatestInfoURL = "https://fias.nalog.ru/WebServices/Public/GetLastDownloadFileInfo"

var entryRe = regexp.MustCompile(`^(\d+)/(AS_[A-Z_]+)_\d{8}_[0-9a-fA-F-]+\.XML$`)

// Archive is an opened gar_xml.zip.
type Archive struct {
	entries map[string]map[string]*zip.File // region -> kind -> file
	remote  *remoteFile                     // nil for a local archive
	closer  io.Closer
}

// LatestFullURL asks the ФНС service for the URL of the latest full export.
func LatestFullURL(ctx context.Context, client *http.Client) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, LatestInfoURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("get latest gar info: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("get latest gar info: unexpected status %s", resp.Status)
	}
	var info struct {
		GarXMLFullURL string `json:"GarXMLFullURL"` //nolint:tagliatelle // ФНС API field name
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", fmt.Errorf("decode latest gar info: %w", err)
	}
	if info.GarXMLFullURL == "" {
		return "", errors.New("latest gar info has no GarXMLFullURL")
	}
	return info.GarXMLFullURL, nil
}

// Open opens a gar_xml.zip from a local path or an http(s) URL.
func Open(ctx context.Context, src string, client *http.Client) (*Archive, error) {
	var (
		ra     io.ReaderAt
		size   int64
		remote *remoteFile
		closer io.Closer
	)
	if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") {
		rf, err := openRemote(ctx, src, client)
		if err != nil {
			return nil, err
		}
		ra, size, remote = rf, rf.size, rf
	} else {
		f, err := os.Open(src)
		if err != nil {
			return nil, fmt.Errorf("open gar archive: %w", err)
		}
		st, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return nil, fmt.Errorf("stat gar archive: %w", err)
		}
		ra, size, closer = f, st.Size(), f
	}

	zr, err := zip.NewReader(ra, size)
	if err != nil {
		if closer != nil {
			_ = closer.Close()
		}
		return nil, fmt.Errorf("read gar archive directory: %w", err)
	}

	a := &Archive{entries: make(map[string]map[string]*zip.File), remote: remote, closer: closer}
	for _, f := range zr.File {
		m := entryRe.FindStringSubmatch(f.Name)
		if m == nil {
			continue
		}
		if a.entries[m[1]] == nil {
			a.entries[m[1]] = make(map[string]*zip.File)
		}
		a.entries[m[1]][m[2]] = f
	}
	if len(a.entries) == 0 {
		_ = a.Close()
		return nil, errors.New("gar archive contains no region folders")
	}
	return a, nil
}

// Close releases the local file, if any.
func (a *Archive) Close() error {
	if a.closer != nil {
		return a.closer.Close()
	}
	return nil
}

// Regions returns the region folder codes ("01" … "99"), sorted.
func (a *Archive) Regions() []string {
	res := make([]string, 0, len(a.entries))
	for code := range a.entries {
		res = append(res, code)
	}
	sort.Strings(res)
	return res
}

// CompressedSize returns the compressed size of a region's file, or 0.
func (a *Archive) CompressedSize(region, kind string) int64 {
	if f := a.entries[region][kind]; f != nil {
		return int64(f.CompressedSize64) //nolint:gosec // zip sizes fit int64
	}
	return 0
}

// OpenFile opens a region's file of the given kind for sequential reading.
// It returns ErrNotFound when the region has no such file.
func (a *Archive) OpenFile(ctx context.Context, region, kind string) (io.ReadCloser, error) {
	f := a.entries[region][kind]
	if f == nil {
		return nil, fmt.Errorf("%s/%s: %w", region, kind, ErrNotFound)
	}
	if a.remote == nil {
		return f.Open()
	}

	// Stream the entry's compressed bytes in one ranged GET instead of
	// letting archive/zip issue small ReadAt calls.
	off, err := f.DataOffset()
	if err != nil {
		return nil, fmt.Errorf("%s/%s: %w", region, kind, err)
	}
	body, err := a.remote.stream(ctx, off, int64(f.CompressedSize64)) //nolint:gosec // zip sizes fit int64
	if err != nil {
		return nil, fmt.Errorf("%s/%s: %w", region, kind, err)
	}
	switch f.Method {
	case zip.Store:
		return body, nil
	case zip.Deflate:
		return &readCloser{Reader: flate.NewReader(body), closers: []io.Closer{body}}, nil
	default:
		_ = body.Close()
		return nil, fmt.Errorf("%s/%s: unsupported zip method %d", region, kind, f.Method)
	}
}

// ErrNotFound is returned by OpenFile for a missing region file.
var ErrNotFound = errors.New("file not found in gar archive")

type readCloser struct {
	io.Reader
	closers []io.Closer
}

func (r *readCloser) Close() error {
	var errs []error
	for _, c := range r.closers {
		errs = append(errs, c.Close())
	}
	if c, ok := r.Reader.(io.Closer); ok {
		errs = append(errs, c.Close())
	}
	return errors.Join(errs...)
}
