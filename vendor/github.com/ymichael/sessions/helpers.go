package sessions

import (
	"bufio"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

var InvalidMessageError = errors.New("Invalid Message.")
var seperator = "."

// Generates random string of length n
func GenerateRandomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func GetSignature(message, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}

// Returns <message>.<signature>
func SignMessage(message, secret string) string {
	return message + seperator + GetSignature(message, secret)
}

// Returns unsigned message.
// If signed message is invalid, returns InvalidMessageError
func UnsignMessage(signedMessage, secret string) (string, error) {
	parts := strings.Split(signedMessage, seperator)
	if len(parts) != 2 {
		return "", InvalidMessageError
	}
	message := parts[0]
	signature := parts[1]
	expectedSignature := GetSignature(message, secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(signature)) {
		return "", InvalidMessageError
	}
	return message, nil
}

// Reverse http.Cookie.String(),
// Taken from: http://play.golang.org/p/YkW_z2CSyE
func CookieFromString(line string) (*http.Cookie, error) {
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(fmt.Sprintf("GET / HTTP/1.0\r\nCookie: %s\r\n\r\n", line))))
	if err != nil {
		return nil, err
	}
	cookies := req.Cookies()
	if len(cookies) == 0 {
		return nil, fmt.Errorf("no cookies")
	}
	return cookies[0], nil
}
