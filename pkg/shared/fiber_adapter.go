package shared

import (
	"log"

	wiring "github.com/4strodev/wiring/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func NewFiberAdapter() *FiberAdapter {
	return new(FiberAdapter)
}

type FiberAdapter struct {
	app *fiber.App
}

func (f *FiberAdapter) Init(container wiring.Container) error {
	f.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	f.app.Use(recover.New())
	f.app.Use(logger.New())

	f.app.Get("/hello", func(ctx *fiber.Ctx) error {
		return ctx.JSON(fiber.Map{
			"message": "Hello world",
		})
	})

	f.app.Hooks().OnListen(func(listenData fiber.ListenData) error {
		if fiber.IsChild() {
			return nil
		}
		log.Printf("listening on port %s:%s\n", listenData.Host, listenData.Port)
		return nil
	})

	container.Singleton(func() fiber.Router {
		return f.app
	})

	return nil
}

func (f *FiberAdapter) Start() error {
	return f.app.Listen(":3000")
}

func (f *FiberAdapter) Stop() error {
	return nil
}
