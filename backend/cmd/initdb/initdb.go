package main

import (
	"log"

	"github.com/kinvolk/nebraska/backend/pkg/api"
)

func main() {
	if _, err := api.New(api.OptionInitDB); err != nil {
		log.Fatal(err)
	}
}
