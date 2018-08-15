package main

import (
	"log"

	"github.com/devigned/buffalo-service-bus-worker/actions"
)

func main() {
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
