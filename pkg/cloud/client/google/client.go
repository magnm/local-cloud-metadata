package google

import (
	"context"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"github.com/magnm/lcm/config"
	"golang.org/x/exp/slog"
	"google.golang.org/api/option"
)

func authentication() option.ClientOption {
	if config.Current.CloudKeyfile != "" {
		return option.WithCredentialsFile(config.Current.CloudKeyfile)
	}
	return nil
}

func GetProject(id string) *resourcemanagerpb.Project {
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

	return project
}
