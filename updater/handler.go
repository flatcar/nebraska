package main

import (
	"context"
	"fmt"
)

type UpdateHandler interface {
	FetchUpdate(ctx context.Context, info *UpdateInfo) error
	ApplyUpdate(ctx context.Context, info *UpdateInfo) error
}

type emptyHandler struct {
}

func NewEmptyHandler() UpdateHandler {
	return emptyHandler{}
}

func (e emptyHandler) FetchUpdate(ctx context.Context, info *UpdateInfo) error {
	fmt.Println("Downloading the upload payload:")
	fmt.Printf("URL: %v\n", info.GetURL())
	fmt.Printf("Version: %v\n", info.GetVersion())
	return nil
}

func (e emptyHandler) ApplyUpdate(ctx context.Context, info *UpdateInfo) error {
	fmt.Println("Installing the update...")
	return nil
}
