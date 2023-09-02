package util

import (
	"net/http"
	"strings"
)

func RequestIp(r *http.Request) string {
	addr := r.RemoteAddr
	return strings.Split(addr, ":")[0]
}
