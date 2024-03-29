package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ea673/voter-api/api"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "8080"
)

func main() {
	app := createApp()

	voterApiHandler := api.NewVoterApi()
	setUpRoutes(app, voterApiHandler)

	host := getEnv("HOST", defaultHost)
	port := getEnv("PORT", defaultPort)
	serverPath := fmt.Sprintf("%s:%s", host, port)

	log.Println("Server is running on", serverPath)
	app.Listen(serverPath)
}

func createApp() *fiber.App {
	app := fiber.New()
	app.Use(cors.New())
	app.Use(recover.New())
	return app
}

func setUpRoutes(app *fiber.App, voterApiHandler *api.VoterApi) {
	app.Get("/voters", voterApiHandler.GetVotersHandler)
	app.Delete("/voters", voterApiHandler.DeleteVotersHandler)

	app.Get("/voters/health", voterApiHandler.GetHealthHandler)

	app.Get("/voters/:id", voterApiHandler.GetVoterHandler)
	app.Post("/voters/:id", voterApiHandler.AddVoterHandler)
	app.Put("/voters/:id", voterApiHandler.UpdateVoterHandler)
	app.Delete("/voters/:id", voterApiHandler.DeleteVoterHandler)

	app.Get("/voters/:id/polls", voterApiHandler.GetVoterHistoriesHandler)

	app.Get("/voters/:id/polls/:pollid", voterApiHandler.GetVoterHistoryHandler)
	app.Post("/voters/:id/polls/:pollid", voterApiHandler.AddVoterHistoryHandler)
	app.Put("/voters/:id/polls/:pollid", voterApiHandler.UpdateVoterHistoryHandler)
	app.Delete("/voters/:id/polls/:pollid", voterApiHandler.DeleteVoterHistoryHandler)

}

func getEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
