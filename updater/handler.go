package main

import (
	"context"
	"fmt"
)

type Handlers interface {
	GetUpdate(ctx context.Context) error
	ApplyUpdate(ctx context.Context) error
}

type emptyHandler struct {
}

func NewEmptyHandler() Handlers {
	return emptyHandler{}
}

func (e emptyHandler) GetUpdate(ctx context.Context) error {
	fmt.Println("Downloading the upload payload")
	return nil
}

func (e emptyHandler) ApplyUpdate(ctx context.Context) error {
	fmt.Println("Installing the update")
	return nil
}
