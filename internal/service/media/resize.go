package media

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/stickpro/go-store/internal/config"
	"github.com/stickpro/go-store/internal/dto"
	"github.com/stickpro/go-store/internal/models"
	"github.com/stickpro/go-store/pkg/imageprocessor"
	"github.com/stickpro/go-store/pkg/object_storage"
)

var (
	// ErrUnknownPreset means the requested size does not match any configured preset.
	ErrUnknownPreset = errors.New("media: unknown image preset")
	// ErrBadImageName means the requested file name is not a safe variant name.
	ErrBadImageName = errors.New("media: invalid image name")
	// ErrOriginalMissing means there is no original for the requested variant.
	ErrOriginalMissing = errors.New("media: original image missing")
	// ErrNotResizable means the original cannot be decoded (corrupt / non-image).
	ErrNotResizable = errors.New("media: image cannot be resized")
)

const imagesDir = "public/images"

// variantNameRe matches "<base>_<size>.<ext>".
var variantNameRe = regexp.MustCompile(`^(.+)_([0-9]{2,5})\.(webp|jpe?g)$`)

// originalExtCandidates is the order in which we look for the original file.
var originalExtCandidates = []string{".jpg", ".jpeg", ".png", ".webp", ".gif"}

type imgPreset struct {
	Name      string
	Size      int
	Fit       imageprocessor.Fit
	NoUpscale bool
}

// PresetNames returns the configured preset names, in order.
func (s *Service) PresetNames() []string {
	return append([]string(nil), s.imgPresetOrder...)
}

// Image builds the ImageDTO for a media record (URLs are pure convention).
func (s *Service) Image(m *models.Medium, alt string) dto.ImageDTO {
	return dto.NewImageDTO(m.ID, m.Path, m.Width, m.Height, alt, s.cfg.Images.ResolvedPresets())
}

// Images maps a gallery to ImageDTOs, all sharing the same alt.
func (s *Service) Images(media []*models.Medium, alt string) []dto.ImageDTO {
	out := make([]dto.ImageDTO, 0, len(media))
	for _, m := range media {
		out = append(out, s.Image(m, alt))
	}
	return out
}

// EnsureImageVariant returns the filesystem path of the "<base>_<size>.<ext>" variant
// under public/images, generating and caching it on the first call. The cache is
// disposable — deleting the "*_<size>.*" files just makes the next request regenerate.
func (s *Service) EnsureImageVariant(ctx context.Context, fileName string) (string, error) {
	m := variantNameRe.FindStringSubmatch(fileName)
	if m == nil {
		return "", ErrBadImageName
	}
	base, sizeStr, ext := m[1], m[2], m[3]
	if !safeImageName(fileName) || !safeImageName(base) {
		return "", ErrBadImageName
	}

	size, _ := strconv.Atoi(sizeStr)
	preset, ok := s.imgPresetsBySize[size]
	if !ok {
		return "", ErrUnknownPreset
	}

	format := imageprocessor.FormatJPEG
	quality := s.imgJPEGQuality
	if ext == "webp" {
		format = imageprocessor.FormatWebP
		quality = s.imgWebPQuality
	}

	variantPath := path.Join(imagesDir, fileName)
	if fsPath, err := s.objectStorage.Get(ctx, variantPath); err == nil {
		return fsPath, nil
	}

	v, err, _ := s.imgGroup.Do(variantPath, func() (any, error) {
		if fsPath, gerr := s.objectStorage.Get(ctx, variantPath); gerr == nil {
			return fsPath, nil
		}

		original, rerr := s.readOriginal(ctx, base)
		if rerr != nil {
			return nil, rerr
		}

		res, perr := s.imgProcessor.Process(original, imageprocessor.Spec{
			Size:      preset.Size,
			Fit:       preset.Fit,
			Format:    format,
			NoUpscale: preset.NoUpscale,
			Quality:   quality,
		})
		if perr != nil {
			if errors.Is(perr, imageprocessor.ErrUnsupportedFormat) || errors.Is(perr, imageprocessor.ErrSourceTooLarge) {
				return nil, fmt.Errorf("%w: %v", ErrNotResizable, perr)
			}
			return nil, fmt.Errorf("resize: %w", perr)
		}

		fsPath, serr := s.objectStorage.Save(ctx, variantPath, res.Data)
		if serr != nil {
			return nil, fmt.Errorf("write variant cache: %w", serr)
		}
		return fsPath, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (s *Service) readOriginal(ctx context.Context, base string) ([]byte, error) {
	for _, ext := range originalExtCandidates {
		data, err := s.objectStorage.Read(ctx, path.Join(imagesDir, base+ext))
		if err == nil {
			return data, nil
		}
		if !errors.Is(err, object_storage.ErrNotFound) {
			return nil, fmt.Errorf("read original: %w", err)
		}
	}
	return nil, ErrOriginalMissing
}

// probeDimensions returns the pixel size of an image, or 0,0 if it can't be read.
func (s *Service) probeDimensions(data []byte) (int32, int32) {
	w, h, err := s.imgProcessor.Probe(data)
	if err != nil {
		return 0, 0
	}
	return int32(w), int32(h) //nolint:gosec
}

// safeImageName rejects path traversal and nested paths.
func safeImageName(name string) bool {
	if name == "" || name == "." || name == ".." {
		return false
	}
	if strings.ContainsAny(name, "/\\") || strings.Contains(name, "..") {
		return false
	}
	return name == path.Base(name)
}

func buildImagePresets(cfg config.ImagesConfig) (byName map[string]imgPreset, bySize map[int]imgPreset, order []string) {
	list := cfg.Presets
	if len(list) == 0 {
		list = config.DefaultImagePresets()
	}
	byName = make(map[string]imgPreset, len(list))
	bySize = make(map[int]imgPreset, len(list))
	for _, p := range list {
		if p.Name == "" || p.Size <= 0 {
			continue
		}
		fit := imageprocessor.FitCover
		if p.Fit == "contain" {
			fit = imageprocessor.FitContain
		}
		pr := imgPreset{Name: p.Name, Size: p.Size, Fit: fit, NoUpscale: p.NoUpscale}
		if _, dup := byName[p.Name]; dup {
			continue
		}
		byName[p.Name] = pr
		bySize[p.Size] = pr
		order = append(order, p.Name)
	}
	return byName, bySize, order
}
