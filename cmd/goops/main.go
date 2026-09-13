package main

import (
	"log"
	"os"

	"github.com/vladimirkorzhenevskiy/goops/internal/app"
)

func main() {
	if err := app.Run(os.Args...); err != nil {
		log.Fatal(err)
	}
}
