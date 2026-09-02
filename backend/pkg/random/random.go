package random

import (
	"crypto/rand"
	"encoding/base32"
)

// This is so we can replace it for testing with something
// deterministic.
var randRead = rand.Read

// DataE returns a slice with n random bytes, or an error if the random source fails.
func DataE(n int) ([]byte, error) {
	if n < 1 {
		return nil, nil
	}
	data := make([]byte, n)
	if _, err := randRead(data); err != nil {
		return nil, err
	}
	return data, nil
}

// Data returns a slice with n random bytes.
func Data(n int) []byte {
	data, err := DataE(n)
	if err != nil {
		// docs say that the error doesn't happen
		panic(err)
	}
	return data
}

// StringE returns a string with n random characters, or an error if the random source fails.
// The characters are based on base32 encoding, containing characters from A-Z and digits from 2 to 7.
func StringE(n int) (string, error) {
	if n < 1 {
		return "", nil
	}
	dataN := ((n-1)/8 + 1) * 5
	data, err := DataE(dataN)
	if err != nil {
		return "", err
	}
	return base32.StdEncoding.EncodeToString(data)[0:n], nil
}

// String returns a string with n random characters. The characters
// are base on base32 encoding, so string will contain characters from
// A-Z and digits from 2 to 7.
func String(n int) string {
	str, err := StringE(n)
	if err != nil {
		panic(err)
	}
	return str
}
