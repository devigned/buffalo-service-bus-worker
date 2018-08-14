package main

import (
	"log"

	"github.com/devigned/demo/actions"
)

func main() {
	app := actions.App()
	if err := app.Serve(); err != nil {
		log.Fatal(err)
	}
}
