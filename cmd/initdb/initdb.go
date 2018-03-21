package main

import (
	"log"

	"github.com/coreroller/coreroller/pkg/api"
)

func main() {
	if _, err := api.New(api.OptionInitDB); err != nil {
		log.Fatal(err)
	}
}
