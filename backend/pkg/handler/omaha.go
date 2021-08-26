package handler

import (
	"bytes"
	"errors"
	"net"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (h *Handler) Omaha(ctx echo.Context) error {
	responseBuffer := new(bytes.Buffer)
	ctx.Response().Writer.Header().Set("Content-Type", "text/xml")
	ctx.Request().Body = http.MaxBytesReader(ctx.Response().Writer, ctx.Request().Body, UpdateMaxRequestSize)
	if err := h.omahaHandler.Handle(ctx.Request().Body, responseBuffer, getRequestIP(ctx.Request())); err != nil {
		logger.Error().Err(err).Msg("process omaha request")
		if uerr := errors.Unwrap(err); uerr != nil && uerr.Error() == "http: request body too large" {
			return ctx.NoContent(http.StatusBadRequest)
		}
	}
	return ctx.XMLBlob(http.StatusOK, responseBuffer.Bytes())
}

func getRequestIP(r *http.Request) string {
	ips := strings.Split(r.Header.Get("X-FORWARDED-FOR"), ",")
	if ips[0] != "" && net.ParseIP(strings.TrimSpace(ips[0])) != nil {
		return ips[0]
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}
