package google

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/magnm/lcm/config"
	googleclient "github.com/magnm/lcm/pkg/cloud/client/google"
	"github.com/magnm/lcm/pkg/kubernetes"
	kubegoogle "github.com/magnm/lcm/pkg/kubernetes/google"
	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

type cachedServiceAccountToken struct {
	Token     string
	ExpiresAt int64
}

var podServiceAccountCache = map[string]string{}
var serviceAccountTokenCache = map[string]cachedServiceAccountToken{}

type recursiveServiceAccountResponse struct {
	Aliases []string `json:"aliases"`
	Email   string   `json:"email"`
	Scopes  []string `json:"scopes"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func serviceAccounts(w http.ResponseWriter, r *http.Request) {
	accountEmail := serviceAccountForPod(w, r)
	if accountEmail == "" {
		slog.Error("no service account found for pod")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if r.URL.Query().Get("recursive") == "true" {
		serviceAccountsRecursive(w, r)
		return
	}

	accounts := []string{
		"default",
		accountEmail,
	}
	accountFolders := lo.Map(accounts, func(acc string, i int) string {
		return acc + "/"
	})
	writeText(w, r, strings.Join(accountFolders, "\n"))
}

func serviceAccountsRecursive(w http.ResponseWriter, r *http.Request) {
	response := map[string]recursiveServiceAccountResponse{
		"default": {
			Aliases: []string{"default"},
			Email:   serviceAccountForPod(w, r),
			Scopes:  googleclient.TokenScopes,
		},
		serviceAccountForPod(w, r): {
			Aliases: []string{"default"},
			Email:   serviceAccountForPod(w, r),
			Scopes:  googleclient.TokenScopes,
		},
	}
	render.JSON(w, r, response)
}

func serviceAccount(w http.ResponseWriter, r *http.Request) {
	accountEmail := serviceAccountForPod(w, r)
	if accountEmail == "" {
		slog.Error("no service account found for pod")
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	if r.URL.Query().Get("recursive") == "true" {
		serviceAccountRecursive(w, r)
		return
	}

	paths := []string{
		"aliases",
		"email",
		"identity",
		"scopes",
		"token",
		"", // Empty line at the end matches GCP behavior
	}
	writeText(w, r, strings.Join(paths, "\n"))
}

func serviceAccountRecursive(w http.ResponseWriter, r *http.Request) {
	response := recursiveServiceAccountResponse{
		Aliases: []string{"default"},
		Email:   serviceAccountForPod(w, r),
		Scopes:  googleclient.TokenScopes,
	}
	render.JSON(w, r, response)
}

func serviceAccountAttr(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "acc")

	// Ensure the email requested is the one currently bound to the pod
	accountEmail := serviceAccountForPod(w, r)
	// If the email is "default", use the account bound to the pod
	if email == "default" {
		email = accountEmail
	}

	if accountEmail == "" || accountEmail != email {
		slog.Error("invalid service account requested", "requested", email, "bound", accountEmail)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	key := chi.URLParam(r, "key")
	slog.Debug("service account attr requested", "email", accountEmail, "key", key)

	switch key {
	case "aliases":
		writeText(w, r, "default")
	case "email":
		writeText(w, r, accountEmail)
	case "identity":
		audience := r.URL.Query().Get("audience")
		if audience == "" {
			http.Error(w, "non-empty audience parameter required", http.StatusBadRequest)
			return
		}
		token := googleclient.GetServiceAccountIdentityToken(accountEmail, audience)
		if token == "" {
			http.Error(w, "failed to get identity token", http.StatusInternalServerError)
			return
		}
		writeText(w, r, token)
	case "scopes":
		writeText(w, r, strings.Join(googleclient.TokenScopes, ","))
	case "token":
		if cached, ok := serviceAccountTokenCache[accountEmail]; ok {
			if cached.ExpiresAt > time.Now().UTC().Unix() {
				render.JSON(w, r, tokenResponse{
					AccessToken: cached.Token,
					ExpiresIn:   int(cached.ExpiresAt - time.Now().UTC().Unix()),
					TokenType:   "Bearer",
				})
				return
			}
		}

		customScopes := strings.Split(r.URL.Query().Get("scopes"), ",")

		token := googleclient.GetServiceAccountToken(accountEmail, customScopes)
		if token == nil {
			http.Error(w, "failed to get access token", http.StatusInternalServerError)
			return
		}
		serviceAccountTokenCache[accountEmail] = cachedServiceAccountToken{
			Token: token.AccessToken,
			// Pretend the token expires 15 minutes before it actually does
			// to avoid caching from returning a just-about-to-expire token
			ExpiresAt: token.ExpiresAt.Add(-15 * time.Minute).Unix(),
		}
		render.JSON(w, r, tokenResponse{
			AccessToken: token.AccessToken,
			ExpiresIn:   int(token.ExpiresAt.Sub(time.Now().UTC()).Seconds()),
			TokenType:   "Bearer",
		})
	}
}

func serviceAccountForPod(w http.ResponseWriter, r *http.Request) string {
	pod, err := kubernetes.CallingPod(r)
	if err != nil {
		slog.Error("failed to get calling pod", "err", err)
		return ""
	}

	if email, ok := podServiceAccountCache[pod.Name]; ok {
		return email
	}

	ksa, err := kubernetes.ServiceAccountForPod(pod)
	if err != nil {
		slog.Error("failed to get service account for pod", "err", err)
		return ""
	}

	var email string

	switch config.Current.KsaResolver {
	case config.KsaBindingResolverAnnotation:
		slog.Debug("using annotation to resolve ksa binding", "ksa", ksa)

		email = kubegoogle.GetGsaForKsa(ksa)
		if email == "" {
			slog.Error("no google service account binding found for ksa", "ksa", ksa)
		}
	case config.KsaBindingResolverCRD:
		slog.Error("using CRD to resolve ksa binding is not implemented", "ksa", ksa)
	}

	if email != "" {
		podServiceAccountCache[pod.Name] = email
	}

	return email
}
