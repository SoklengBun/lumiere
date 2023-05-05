package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

func main(){
	app := fiber.New()
	app.Get("/api", welcome)
	app.Post("/api/login", login)

	log.Fatal(app.Listen(":3000"))
}

func welcome(c *fiber.Ctx) error {
	return c.SendString("Hey Hey! Shinomiya San")
}

func login(c *fiber.Ctx) error {
	return c.SendString("Hey Hey! Shinomiya San login")
}