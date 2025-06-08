package routes

import (
	"backend/core/handlers"
	"backend/core/middlewares/validators"

	"github.com/gofiber/fiber/v3"
)

func SetupAuthRoutes(app *fiber.App) {
	group := app.Group("/auth")

	group.Post("/register", handlers.Register, validators.ValidateRegisterInfo)
	group.Post("/login", handlers.Login, validators.ValidateLoginInfo)
	group.Post("/token/verify", handlers.ValidateToken)
	group.Post("/token/refresh", handlers.RefreshAccessToken)
}
