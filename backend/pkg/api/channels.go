package api

import "github.com/flatcar/nebraska/backend/pkg/api/internal/types"

var (
	// ErrInvalidPackage error indicates that a package doesn't belong to the
	// application it was supposed to belong to.
	ErrInvalidPackage = types.ErrInvalidPackage

	// ErrBlacklistedChannel error indicates an attempt of creating/updating a
	// channel using a package that has blacklisted the channel.
	ErrBlacklistedChannel = types.ErrBlacklistedChannel
)

type Channel = types.Channel
