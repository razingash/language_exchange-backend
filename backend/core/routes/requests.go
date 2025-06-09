package routes

import (
	"backend/core/handlers"
	"backend/core/middlewares"
	"backend/core/middlewares/validators"

	"github.com/gofiber/fiber/v3"
)

func SetupRequestsRoutes(app *fiber.App) {
	group := app.Group("/requests")

	group.Post("/", handlers.CreateMatchRequest, middlewares.IsAuthorized, validators.ValidatePostMatchRequest)
	group.Get("/incoming", handlers.GetIncomingMatchRequest, middlewares.IsAuthorized)
	group.Get("/outcoming", handlers.GetOutgoingMatchRequest, middlewares.IsAuthorized)
	group.Get("/matches/", handlers.GetAcceptedMatchRequest, middlewares.IsAuthorized)
	group.Put("/:id/accept", handlers.PutAcceptMatchRequest, middlewares.IsAuthorized, validators.ValidateMatchOwnership)
	group.Put("/:id/decline", handlers.PutDeclineMatchRequest, middlewares.IsAuthorized, validators.ValidateMatchOwnership)
}
