package auth

import (
	wiring "github.com/4strodev/wiring/pkg"
	"github.com/gofiber/fiber/v2"
)

type AuthController struct {
}

func NewAuthController() *AuthController {
	return &AuthController{}
}

func (c *AuthController) Init(container wiring.Container) error {
	var router fiber.Router
	err := container.Resolve(&router)
	if err != nil {
		return err
	}
	router.Get("/login", c.Login)
	return nil
}

func (c *AuthController) Login(ctx *fiber.Ctx) error {
	return ctx.JSON(fiber.Map{
		"message": "Login will be implented",
	})
}
