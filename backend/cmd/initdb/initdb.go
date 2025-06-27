package main

import (
	"log"

	"github.com/flatcar/nebraska/backend/pkg/api"
)

func main() {
	if _, err := api.NewWithMigrations(api.OptionInitDB); err != nil {
		log.Fatal(err)
	}
}
