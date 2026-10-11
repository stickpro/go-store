package gar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// remoteBlockSize is the granularity of cached ReadAt fetches. archive/zip
// reads the central directory (~200 KB for gar_xml.zip) in small chunks.
const remoteBlockSize = 1 << 20

// remoteFile is an io.ReaderAt over an HTTP resource that supports ranges.
type remoteFile struct {
	ctx    context.Context //nolint:containedctx // io.ReaderAt has no ctx parameter
	url    string
	size   int64
	client *http.Client

	mu     sync.Mutex
	blocks map[int64][]byte
}

func openRemote(ctx context.Context, url string, client *http.Client) (*remoteFile, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("head %s: %w", url, err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("head %s: unexpected status %s", url, resp.Status)
	}
	if resp.Header.Get("Accept-Ranges") != "bytes" || resp.ContentLength <= 0 {
		return nil, fmt.Errorf("%s does not support range requests; download it and pass the local path", url)
	}
	return &remoteFile{
		ctx:    ctx,
		url:    url,
		size:   resp.ContentLength,
		client: client,
		blocks: make(map[int64][]byte),
	}, nil
}

func (r *remoteFile) ReadAt(p []byte, off int64) (int, error) {
	if off >= r.size {
		return 0, io.EOF
	}
	n := 0
	for n < len(p) && off+int64(n) < r.size {
		pos := off + int64(n)
		block, err := r.block(pos / remoteBlockSize)
		if err != nil {
			return n, err
		}
		n += copy(p[n:], block[pos%remoteBlockSize:])
	}
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (r *remoteFile) block(idx int64) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.blocks[idx]; ok {
		return b, nil
	}
	start := idx * remoteBlockSize
	length := min(int64(remoteBlockSize), r.size-start)
	body, err := r.stream(r.ctx, start, length)
	if err != nil {
		return nil, err
	}
	defer func() { _ = body.Close() }()
	b := make([]byte, length)
	if _, err := io.ReadFull(body, b); err != nil {
		return nil, fmt.Errorf("read %s bytes %d-%d: %w", r.url, start, start+length-1, err)
	}
	r.blocks[idx] = b
	return b, nil
}

// stream opens bytes [off, off+length) of the resource for sequential reading.
func (r *remoteFile) stream(ctx context.Context, off, length int64) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", off, off+length-1))
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get %s: %w", r.url, err)
	}
	if resp.StatusCode != http.StatusPartialContent {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("get %s bytes %d-%d: unexpected status %s", r.url, off, off+length-1, resp.Status)
	}
	return resp.Body, nil
}
