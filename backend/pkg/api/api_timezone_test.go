package api

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithUTCSessionTimezone(t *testing.T) {
	t.Parallel()

	out := withUTCSessionTimezone("postgres://u:p@localhost/db?sslmode=disable")
	u, err := url.Parse(out)
	assert.NoError(t, err)
	assert.Equal(t, "UTC", u.Query().Get("TimeZone"))
	assert.Equal(t, "disable", u.Query().Get("sslmode"))

	// Preserve an explicit TimeZone already set by the operator.
	in := "postgres://u:p@localhost/db?TimeZone=Asia%2FKolkata"
	assert.Equal(t, in, withUTCSessionTimezone(in))

	// Preserve timezone=... (lowercase) as well.
	in = "postgres://u:p@localhost/db?timezone=Europe/Berlin"
	assert.Equal(t, in, withUTCSessionTimezone(in))

	// Preserve options=-c TimeZone=... overrides.
	in = "postgres://u:p@localhost/db?options=-c%20TimeZone%3DUTC"
	assert.Equal(t, in, withUTCSessionTimezone(in))
}
