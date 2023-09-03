package google

import (
	"context"
	"errors"
	"strings"
	"time"

	"cloud.google.com/go/iam"
	iamadmin "cloud.google.com/go/iam/admin/apiv1"
	iampb "cloud.google.com/go/iam/apiv1/iampb"
	iamcredentials "cloud.google.com/go/iam/credentials/apiv1"
	iamcredentialspb "cloud.google.com/go/iam/credentials/apiv1/credentialspb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	durationpb "github.com/golang/protobuf/ptypes/duration"
	"github.com/magnm/lcm/config"
	"golang.org/x/exp/slog"
	google "golang.org/x/oauth2/google"
	"google.golang.org/api/oauth2/v1"
	"google.golang.org/api/option"
)

type Token struct {
	AccessToken string
	ExpiresAt   time.Time
}

var TokenScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
}

var IdentityWorkloadRole = "roles/iam.workloadIdentityUser"
var TokenCreatorRole = "roles/iam.serviceAccountTokenCreator"

var cachedProject *resourcemanagerpb.Project
var serviceAccountPermissionCache = map[string]bool{}

func GetProject(id string) *resourcemanagerpb.Project {
	if cachedProject != nil {
		return cachedProject
	}

	slog.Debug("getting google project", "id", id)
	ctx := context.Background()
	client, err := resourcemanager.NewProjectsClient(ctx, authentication())
	if err != nil {
		slog.Error("failed to create resourcemanager client", "err", err)
		return nil
	}
	defer client.Close()

	projectIterator := client.SearchProjects(ctx, &resourcemanagerpb.SearchProjectsRequest{
		Query: "id:" + id,
	})

	cachedProject, _ = projectIterator.Next()
	if cachedProject == nil {
		slog.Error("failed to get project", "id", id)
	} else {
		slog.Debug("got project", "id", id, "name", cachedProject.Name)
	}

	return cachedProject
}

func ValidateKsaGsaBinding(ksaBinding string, gsa string) bool {
	slog.Debug("validating ksa gsa binding", "ksaBinding", ksaBinding, "gsa", gsa)
	ctx := context.Background()

	client, err := iamadmin.NewIamClient(ctx, authentication())
	if err != nil {
		slog.Error("failed to create iamadmin client", "err", err)
		return false
	}
	defer client.Close()

	policy, err := client.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: "projects/-/serviceAccounts/" + gsa,
	})
	if err != nil {
		slog.Error("failed to get iam policy", "err", err)
		return false
	}

	for _, member := range policy.Members(iam.RoleName(IdentityWorkloadRole)) {
		if member == ksaBinding {
			slog.Debug("validated ksa gsa binding", "ksaBinding", ksaBinding, "gsa", gsa)
			return true
		}
	}

	slog.Warn("invalid ksa gsa binding", "ksaBinding", ksaBinding, "gsa", gsa)

	return false
}

func GetMainAccount() string {
	slog.Debug("getting main account")
	ctx := context.Background()
	client, err := oauth2.NewService(ctx, authentication())
	if err != nil {
		slog.Error("failed to create oauth2 client", "err", err)
		return ""
	}
	resp, err := client.Tokeninfo().Do()
	if err != nil {
		slog.Error("failed to get userinfo", "err", err)
		return ""
	}
	return resp.Email
}

func GetMainAccountAccessToken() string {
	slog.Debug("getting main account access token")
	ctx := context.Background()
	credentials, err := google.FindDefaultCredentials(ctx, TokenScopes...)
	if err != nil {
		slog.Error("failed to get default credentials", "err", err)
		return ""
	}
	token, err := credentials.TokenSource.Token()
	if err != nil {
		slog.Error("failed to get credentials token", "err", err)
		return ""
	}
	return token.AccessToken
}

func GetServiceAccountToken(email string, scopes []string) *Token {
	slog.Debug("getting service account token", "email", email)
	ctx := context.Background()

	// Make sure we are allowed to generate tokens
	if !verifyTokenCreatorOnServiceAccount(email) {
		if err := selfGrantTokenCreatorOnServiceAccount(email); err != nil {
			slog.Error("failed to grant token creator role on service account", "err", err)
			return nil
		}
	}

	client, err := iamcredentials.NewIamCredentialsClient(ctx, authentication())
	if err != nil {
		slog.Error("failed to create iamcredentials client", "err", err)
		return nil
	}
	defer client.Close()

	if len(scopes) == 0 {
		scopes = TokenScopes
	}

	token, err := client.GenerateAccessToken(ctx, &iamcredentialspb.GenerateAccessTokenRequest{
		Name:  "projects/-/serviceAccounts/" + email,
		Scope: scopes,
		Lifetime: &durationpb.Duration{
			Seconds: 3600,
		},
	})
	if err != nil {
		slog.Error("failed to get access token", "err", err)
		return nil
	}

	slog.Debug("got token", "email", email)

	return &Token{
		AccessToken: token.AccessToken,
		ExpiresAt:   token.ExpireTime.AsTime().UTC(),
	}
}

func GetServiceAccountIdentityToken(email string, audience string) string {
	slog.Debug("getting service account identity token", "email", email, "audience", audience)
	ctx := context.Background()

	// Make sure we are allowed to generate tokens
	if !verifyTokenCreatorOnServiceAccount(email) {
		if err := selfGrantTokenCreatorOnServiceAccount(email); err != nil {
			slog.Error("failed to grant token creator role on service account", "err", err)
			return ""
		}
	}

	client, err := iamcredentials.NewIamCredentialsClient(ctx, authentication())
	if err != nil {
		slog.Error("failed to create iamcredentials client", "err", err)
		return ""
	}
	defer client.Close()

	token, err := client.GenerateIdToken(ctx, &iamcredentialspb.GenerateIdTokenRequest{
		Name:     "projects/-/serviceAccounts/" + email,
		Audience: audience,
	})
	if err != nil {
		slog.Error("failed to get identity token", "err", err)
		return ""
	}

	slog.Debug("got identity token", "email", email, "audience", audience)

	return token.Token
}

func authentication() option.ClientOption {
	if config.Current.CloudKeyfile != "" {
		return option.WithCredentialsFile(config.Current.CloudKeyfile)
	}
	return option.WithTelemetryDisabled()
}

func verifyTokenCreatorOnServiceAccount(email string) bool {
	if val, ok := serviceAccountPermissionCache[email]; ok {
		return val
	}

	slog.Debug("verifying token creator role on service account", "email", email)
	ctx := context.Background()

	client, err := iamadmin.NewIamClient(ctx, authentication())
	if err != nil {
		slog.Error("failed to create iam client", "err", err)
		return false
	}
	defer client.Close()

	permissions, err := client.TestIamPermissions(ctx, &iampb.TestIamPermissionsRequest{
		Resource: "projects/-/serviceAccounts/" + email,
		Permissions: []string{
			"iam.serviceAccounts.getAccessToken",
		},
	})
	if err != nil {
		slog.Error("failed to test iam permissions", "err", err)
		return false
	}

	if len(permissions.Permissions) == 0 {
		slog.Warn("token creator role not granted on service account", "email", email)
		return false
	}

	slog.Debug("verified token creator role on service account", "email", email)
	serviceAccountPermissionCache[email] = true

	return true
}

func selfGrantTokenCreatorOnServiceAccount(email string) error {
	slog.Debug("granting token creator role on service account", "email", email)
	ctx := context.Background()

	client, err := iamadmin.NewIamClient(ctx, authentication())
	if err != nil {
		slog.Error("failed to create iam client", "err", err)
		return err
	}
	defer client.Close()

	existingPolicy, err := client.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: "projects/-/serviceAccounts/" + email,
	})
	if err != nil {
		slog.Error("failed to get existing iam policy", "err", err)
	}

	mainAccount := GetMainAccount()
	if mainAccount == "" {
		return errors.New("failed to get main account")
	}

	principal := mainAccount
	if strings.HasSuffix(principal, "gserviceaccount.com") {
		principal = "serviceAccount:" + principal
	} else {
		principal = "user:" + principal
	}

	slog.Debug("adding token creator role to main account", "principal", principal)
	existingPolicy.Add(principal, iam.RoleName(TokenCreatorRole))

	_, err = client.SetIamPolicy(ctx, &iamadmin.SetIamPolicyRequest{
		Resource: "projects/-/serviceAccounts/" + email,
		Policy:   existingPolicy,
	})
	if err != nil {
		slog.Error("failed to set iam policy", "err", err)
		return err
	}

	slog.Debug("granted token creator role on service account", "email", email, "principal", principal)
	serviceAccountPermissionCache[email] = true
	return nil
}
