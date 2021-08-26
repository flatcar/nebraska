package main

import (
	"context"
	"fmt"
)

type emptyHandler struct {
}

func NewEmptyHandler() UpdateHandler {
	return emptyHandler{}
}

func (e emptyHandler) FetchUpdate(ctx context.Context) error {
	fmt.Println("Downloading the upload payload")
	return nil
}

func (e emptyHandler) ApplyUpdate(ctx context.Context) error {
	fmt.Println("Installing the update")
	return nil
}
