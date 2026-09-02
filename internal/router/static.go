package router

import (
	"errors"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/stickpro/go-store/internal/service/media"
)

func (r *Router) initStaticFiles(app *fiber.App) {
	app.Get("/swagger.yaml", func(c fiber.Ctx) error {
		return c.SendFile("docs/swagger.yaml")
	})

	// Resized variants live next to the originals as "<base>_<size>.<webp|jpg>".
	// Names that don't look like a variant fall through to the static handler.
	// Registered before it so it wins over the /storage/public* wildcard.
	app.Get("/storage/public/images/:file", r.serveImageVariant)

	app.Get("/storage/public*", static.New("", static.Config{
		FS:     os.DirFS(r.config.FileStorage.Path + "/public"),
		Browse: true,
	}))
}

func (r *Router) serveImageVariant(c fiber.Ctx) error {
	fsPath, err := r.services.MediaService.EnsureImageVariant(c.Context(), c.Params("file"))
	if err != nil {
		if errors.Is(err, media.ErrBadImageName) {
			// not a variant name (e.g. an original) -> let the static handler serve it
			return c.Next()
		}
		switch {
		case errors.Is(err, media.ErrUnknownPreset),
			errors.Is(err, media.ErrOriginalMissing),
			errors.Is(err, media.ErrNotResizable):
			return fiber.ErrNotFound
		default:
			r.logger.Error("failed to build image variant", "file", c.Params("file"), "error", err)
			return fiber.ErrInternalServerError
		}
	}

	c.Set(fiber.HeaderCacheControl, "public, max-age=31536000, immutable")
	return c.SendFile(fsPath)
}
