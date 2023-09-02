package google

import (
	"context"

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
	"google.golang.org/api/option"
)

var TokenScopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
}

var IdentityWorkloadRole = "roles/iam.workloadIdentityUser"

func authentication() option.ClientOption {
	if config.Current.CloudKeyfile != "" {
		return option.WithCredentialsFile(config.Current.CloudKeyfile)
	}
	return nil
}

func GetProject(id string) *resourcemanagerpb.Project {
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

	project, _ := projectIterator.Next()
	if project == nil {
		slog.Error("failed to get project", "id", id)
	} else {
		slog.Debug("got project", "id", id, "name", project.Name)
	}

	return project
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

func GetServiceAccountAccessToken(email string) string {
	slog.Debug("getting service account access token", "email", email)
	ctx := context.Background()

	client, err := iamcredentials.NewIamCredentialsClient(ctx, authentication())
	if err != nil {
		slog.Error("failed to create iamcredentials client", "err", err)
		return ""
	}
	defer client.Close()

	token, err := client.GenerateAccessToken(ctx, &iamcredentialspb.GenerateAccessTokenRequest{
		Name:  "projects/-/serviceAccounts/" + email,
		Scope: TokenScopes,
		Lifetime: &durationpb.Duration{
			Seconds: 3600,
		},
	})
	if err != nil {
		slog.Error("failed to get access token", "err", err)
		return ""
	}

	slog.Debug("got access token", "email", email)

	return token.AccessToken
}

func GetServiceAccountIdentityToken(email string, audience string) string {
	slog.Debug("getting service account identity token", "email", email, "audience", audience)
	ctx := context.Background()

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
