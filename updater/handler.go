package main

import (
	"fmt"
)

type emptyHandler struct {
}

func NewEmptyHandler() UpdateHandler {
	return emptyHandler{}
}

func (e emptyHandler) FetchUpdate(info *UpdateInfo) error {
	fmt.Println("Downloading the upload payload:")
	fmt.Printf("URL: %v\n", info.GetURL())
	fmt.Printf("Version: %v\n", info.GetVersion())
	return nil
}

func (e emptyHandler) ApplyUpdate(info *UpdateInfo) error {
	fmt.Println("Installing the update...")
	return nil
}
