package random

import (
	"crypto/rand"
	"encoding/base32"
)

// This is so we can replace it for testing with something
// deterministic.
var randRead func([]byte) (int, error) = rand.Read

// Data returns a slice with n random bytes.
func Data(n int) []byte {
	if n < 1 {
		return nil
	}
	data := make([]byte, n)
	if _, err := randRead(data); err != nil {
		// docs say that the error doesn't happen
		panic(err)
	}
	return data
}

// String returns a string with n random characters. The characters
// are base on base32 encoding, so string will contain characters from
// A-Z and digits from 2 to 7.
func String(n int) string {
	if n < 1 {
		return ""
	}
	dataN := ((n-1)/8 + 1) * 5
	return base32.StdEncoding.EncodeToString(Data(dataN))[0:n]
}
