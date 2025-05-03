package httpapi

import (
	"github.com/twigman/fshare/src/config"
)

type HTTPHandler struct {
	config *config.Config
	reg    *config.FileRegister
}

func NewHTTPHandler(config *config.Config) *HTTPHandler {
	return &HTTPHandler{config: config}
}
