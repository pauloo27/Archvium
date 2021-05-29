package router

import (
	authController "github.com/Pauloo27/archvium/controllers/auth"
	usersController "github.com/Pauloo27/archvium/controllers/users"
	"github.com/gofiber/fiber/v2"
)

func RouteFor(app *fiber.App) {

	v1 := app.Group("/v1")
	auth := v1.Group("/auth")
	users := v1.Group("/users")

	auth.Post("/register",
		requireGuest,
		withEnvBool("AUTH_SELF_REGISTER"),
		withPayload(new(authController.RegisterPayload)),
		authController.Register,
	)
	auth.Post("/login",
		requireGuest,
		withPayload(new(authController.LoginPayload)),
		authController.Login,
	)

	users.Get("/@me", requireAuth, usersController.GetMe)
}
