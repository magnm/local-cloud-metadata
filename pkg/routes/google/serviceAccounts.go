package google

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi"
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
	writeText(w, r, strings.Join(accountFolders, "\n"))
}

func serviceAccount(w http.ResponseWriter, r *http.Request) {
	paths := []string{
		"aliases",
		"email",
		"identity",
		"scopes",
		"token",
	}
	writeText(w, r, strings.Join(paths, "\n"))
}

func serviceAccountProp(w http.ResponseWriter, r *http.Request) {
	switch chi.URLParam(r, "key") {
	case "aliases":
		writeText(w, r, "default")
	case "email":
		writeText(w, r, "account email")
	case "identity":
		audience := r.URL.Query().Get("audience")
		if audience == "" {
			http.Error(w, "non-empty audience parameter required", http.StatusBadRequest)
			return
		}
		writeText(w, r, "token-using-audience-here")
	case "scopes":
		writeText(w, r, "https://www.googleapis.com/auth/cloud-platform")
	case "token":
		writeText(w, r, "token-here")
	}
}
