package routes

import (
	"backend/core/handlers"
	"backend/core/middlewares"
	"backend/core/middlewares/validators"

	"github.com/gofiber/fiber/v3"
)

func SetupUsersRoutes(app *fiber.App) {
	group := app.Group("/users")

	group.Get("/", handlers.GetTargetedUsers, middlewares.IsAuthorized)
	group.Get("/me", handlers.GetUserInfo, middlewares.IsAuthorized)
	group.Put("/me/languages", handlers.UpdateUserLanguages, middlewares.IsAuthorized, validators.ValidateLanguages)
}
