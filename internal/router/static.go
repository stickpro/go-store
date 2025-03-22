package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"
	"os"
)

var StaticDir = os.DirFS("./storage/")

func (r *Router) initStaticFiles(app *fiber.App) {
	app.Get("/swagger.yaml", func(c fiber.Ctx) error {
		return c.SendFile("docs/swagger.yaml")
	})

	app.Get("/storage/public*", static.New("", static.Config{
		FS:     os.DirFS(r.config.FileStorage.Path + "/public"),
		Browse: true,
	}))

}
