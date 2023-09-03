package google

import (
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/magnm/lcm/config"
	googleclient "github.com/magnm/lcm/pkg/cloud/client/google"
	"github.com/magnm/lcm/pkg/kubernetes"
	kubegoogle "github.com/magnm/lcm/pkg/kubernetes/google"
	"github.com/samber/lo"
	"golang.org/x/exp/slog"
)

type cachedServiceAccountToken struct {
	token     string
	expiresAt int64
}

var podServiceAccountCache = map[string]string{}
var serviceAccountTokenCache = map[string]cachedServiceAccountToken{}

func serviceAccounts(w http.ResponseWriter, r *http.Request) {
	accountEmail := serviceAccountForPod(w, r)
	if accountEmail == "" {
		slog.Error("no service account found for pod")
		http.Error(w, "no service account found for pod", http.StatusInternalServerError)
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
			if cached.expiresAt > time.Now().UTC().Unix() {
				writeText(w, r, cached.token)
				return
			}
		}

		token := googleclient.GetServiceAccountToken(accountEmail)
		if token == nil {
			http.Error(w, "failed to get access token", http.StatusInternalServerError)
			return
		}
		serviceAccountTokenCache[accountEmail] = cachedServiceAccountToken{
			token:     token.AccessToken,
			expiresAt: token.ExpiresAt.Unix(),
		}
		writeText(w, r, token.AccessToken)
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
