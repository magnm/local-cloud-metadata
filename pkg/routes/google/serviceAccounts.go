package google

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/samber/lo"
)

func serviceAccounts(w http.ResponseWriter, r *http.Request) {
	accounts := []string{
		"default",
		"another",
	}
	accountFolders := lo.Map(accounts, func(acc string, i int) string {
		return acc + "/"
	})
	render.PlainText(w, r, strings.Join(accountFolders, "\n"))
}

func serviceAccount(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		"aliases",
		"email",
		"identity",
		"scopes",
		"token",
	}
	render.PlainText(w, r, strings.Join(paths, "\n"))
}

func serviceAccountProp(w http.ResponseWriter, r *http.Request) {
	switch chi.URLParam(r, "key") {
	case "aliases":
		render.PlainText(w, r, "")
	case "email":
		render.PlainText(w, r, "")
	case "identity":
		render.PlainText(w, r, "")
	case "scopes":
		render.PlainText(w, r, "")
	case "token":
		render.PlainText(w, r, "")
	}
}
