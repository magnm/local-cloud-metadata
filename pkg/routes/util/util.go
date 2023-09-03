package util

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/samber/lo"
)

func RedirectTo(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		patterns := lo.Map(chi.RouteContext(r.Context()).RoutePatterns, func(pattern string, i int) string {
			return strings.ReplaceAll(pattern, "/*", "")
		})
		patterns = lo.Filter(patterns, func(pattern string, i int) bool {
			return pattern != ""
		})
		initialPath := strings.Join(lo.Slice(patterns, 0, len(patterns)-1), "/")
		http.Redirect(w, r, initialPath+path, http.StatusPermanentRedirect)
	}
}

func RedirectToKey(path string, keys []string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resolvedPath := path
		for _, key := range keys {
			value := chi.URLParam(r, key)
			resolvedPath = strings.ReplaceAll(resolvedPath, fmt.Sprintf("{%s}", key), value)
		}
		patterns := lo.Map(chi.RouteContext(r.Context()).RoutePatterns, func(pattern string, i int) string {
			return strings.ReplaceAll(pattern, "/*", "")
		})
		patterns = lo.Filter(patterns, func(pattern string, i int) bool {
			return pattern != ""
		})
		initialPath := strings.Join(lo.Slice(patterns, 0, len(patterns)-1), "/")
		http.Redirect(w, r, initialPath+resolvedPath, http.StatusPermanentRedirect)
	}
}
