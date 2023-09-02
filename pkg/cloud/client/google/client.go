package google

import (
	"context"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"golang.org/x/exp/slog"
)

func GetProject(id string) *resourcemanagerpb.Project {
	ctx := context.Background()
	client, err := resourcemanager.NewProjectsClient(ctx)
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
