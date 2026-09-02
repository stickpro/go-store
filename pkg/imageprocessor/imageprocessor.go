// Package imageprocessor resizes and re-encodes images via libvips (cgo).
package imageprocessor

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"strings"
	"sync"

	"github.com/davidbyttow/govips/v2/vips"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

// Fit controls how the source is mapped onto the target box.
type Fit string

const (
	// FitCover scales and center-crops so the result fills the box.
	FitCover Fit = "cover"
	// FitContain scales so the whole image fits inside the box.
	FitContain Fit = "contain"
)

// Format is an output encoding.
type Format string

const (
	FormatJPEG Format = "jpeg"
	FormatWebP Format = "webp"
)

var (
	// ErrUnsupportedFormat is returned when the source cannot be decoded.
	ErrUnsupportedFormat = errors.New("imageprocessor: unsupported image format")
	// ErrSourceTooLarge is returned when the source pixel count exceeds the limit.
	ErrSourceTooLarge = errors.New("imageprocessor: source image too large")
)

// Spec describes one resize+encode operation. The target box is Size x Size.
type Spec struct {
	Size      int
	Fit       Fit
	Format    Format
	NoUpscale bool
	Quality   int
}

// Result is a re-encoded image variant.
type Result struct {
	Data        []byte
	ContentType string
	Width       int
	Height      int
}

// Processor resizes images.
type Processor interface {
	// Process decodes src, resizes per spec and re-encodes, dropping metadata.
	Process(src []byte, spec Spec) (Result, error)
	// Probe returns the source pixel dimensions without a full decode.
	Probe(src []byte) (width, height int, err error)
}

var startup sync.Once

// Startup boots libvips. Safe to call multiple times; the first call wins.
func Startup() {
	startup.Do(func() {
		vips.LoggingSettings(nil, vips.LogLevelError)
		vips.Startup(nil)
	})
}

// Shutdown releases libvips resources. Call once on process exit.
func Shutdown() { vips.Shutdown() }

// New returns a libvips-backed processor. maxSourcePixels caps the source area
// (width*height) as a decompression-bomb guard; 0 disables the check.
func New(maxSourcePixels int) Processor {
	Startup()
	return &processor{maxSourcePixels: maxSourcePixels}
}

type processor struct {
	maxSourcePixels int
}

// trimVipsStack drops the goroutine dump govips appends to every error message.
func trimVipsStack(err error) string {
	s := err.Error()
	if i := strings.Index(s, "\nStack:"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return s
}

func (p *processor) Probe(src []byte) (int, int, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(src))
	if err != nil {
		return 0, 0, fmt.Errorf("%w: %v", ErrUnsupportedFormat, err)
	}
	return cfg.Width, cfg.Height, nil
}

func (p *processor) Process(src []byte, spec Spec) (Result, error) {
	if p.maxSourcePixels > 0 {
		if w, h, err := p.Probe(src); err == nil && w*h > p.maxSourcePixels {
			return Result{}, fmt.Errorf("%w: %dx%d", ErrSourceTooLarge, w, h)
		}
	}

	crop := vips.InterestingNone
	if spec.Fit == FitCover {
		crop = vips.InterestingCentre
	}
	size := vips.SizeBoth
	if spec.NoUpscale {
		size = vips.SizeDown
	}

	img, err := vips.NewThumbnailWithSizeFromBuffer(src, spec.Size, spec.Size, crop, size)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %s", ErrUnsupportedFormat, trimVipsStack(err))
	}
	defer img.Close()

	q := spec.Quality
	if q <= 0 || q > 100 {
		q = 82
	}

	var (
		data        []byte
		contentType string
	)
	switch spec.Format {
	case FormatWebP:
		data, _, err = img.ExportWebp(&vips.WebpExportParams{Quality: q, StripMetadata: true})
		contentType = "image/webp"
	default:
		data, _, err = img.ExportJpeg(&vips.JpegExportParams{Quality: q, StripMetadata: true, OptimizeCoding: true})
		contentType = "image/jpeg"
	}
	if err != nil {
		return Result{}, fmt.Errorf("imageprocessor: encode %s: %s", spec.Format, trimVipsStack(err))
	}

	return Result{
		Data:        data,
		ContentType: contentType,
		Width:       img.Width(),
		Height:      img.Height(),
	}, nil
}
