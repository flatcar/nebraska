package sessions

import (
	"net/http"
	"time"
)

// Options used to create http.Cookie
type CookieOptions struct {
	Path     string
	MaxAge   int  // Setting to less than 0 deletes cookie
	HttpOnly bool // Prevents JavaScript access to cookies
	Secure   bool // Cookie wil only be transmitted over SSL/TLS.
}

// NewCookie returns an http.Cookie with the options set. It also sets
// the Expires field calculated based on the MaxAge value, for Internet
// Explorer compatibility.
// Adapted from gorilla/sessions
func NewCookie(name, value string, options *CookieOptions) *http.Cookie {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     options.Path,
		MaxAge:   options.MaxAge,
		Secure:   options.Secure,
		HttpOnly: options.HttpOnly,
	}
	if options.MaxAge > 0 {
		d := time.Duration(options.MaxAge) * time.Second
		cookie.Expires = time.Now().Add(d)
	} else if options.MaxAge < 0 {
		// Set it to the past to expire now.
		cookie.Expires = time.Unix(1, 0)
	}
	return cookie
}
