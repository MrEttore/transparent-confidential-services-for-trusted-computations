package workload

import (
	"context"
	"fmt"
	"log"

	"evidence-provider/internal/service/mock"
	"evidence-provider/internal/types"
	"evidence-provider/internal/util"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// FetchEvidence gathers evidence from running containers and their associated images.
//
// If MOCK_MODE environment variable is set, returns sample mock data for local development.
func FetchEvidence(ctx context.Context, challenge string) (types.WorkloadEvidence, error) {
	if mock.IsMockMode() {
		log.Println("Mock mode enabled: returning sample workload evidence")

		return types.WorkloadEvidence{
			Containers: []types.ContainerEvidence{
				{
					ID:          "90f62b5f43c6a1c06e6df91d97476fbf6d624e30663c4fabf3400ef40886cbe7",
					Name:        "llm-core",
					Image:       "sanctuairy/llm-core:latest",
					ImageDigest: "sha256:435faa6db70075a575dd54e2a1e76cee14bd53fb67be4fdfa3736d879b9f1ccb",
					State:       "running",
					StartedAt:   "2025-12-08T13:41:20.710779927Z",
					Labels: map[string]string{
						"org.opencontainers.image.ref.name": "ubuntu",
						"org.opencontainers.image.version":  "24.04",
					},
				},
			},
			Images: []types.ImageEvidence{
				{
					ID:       "sha256:435faa6db70075a575dd54e2a1e76cee14bd53fb67be4fdfa3736d879b9f1ccb",
					RepoTags: []string{"sanctuairy/llm-core:latest"},
					RepoDigests: []string{
						"sanctuairy/llm-core@sha256:1fc562c7ec052c28bd87c1e0d788a2f542fd177559e78a5db9cfd258db0b3d51",
					},
					Created: "2025-11-13T13:35:24.498087954Z",
					Size:    3653158191,
					Labels: map[string]string{
						"org.opencontainers.image.ref.name": "ubuntu",
						"org.opencontainers.image.version":  "24.04",
					},
				},
			},
			ReportData: challenge,
		}, nil
	}

	cli, err := client.NewClientWithOpts(
		client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return types.WorkloadEvidence{}, fmt.Errorf("failed to create a new client: %w", err)
	}

	containerList, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return types.WorkloadEvidence{}, fmt.Errorf("failed to list containers: %w", err)
	}

	var containerEvidence []types.ContainerEvidence
	var imagesSeen = make(map[string]struct{})
	var imageEvidence []types.ImageEvidence

	for _, c := range containerList {
		ci, _ := cli.ContainerInspect(ctx, c.ID)
		containerEvidence = append(containerEvidence, types.ContainerEvidence{
			ID:          c.ID,
			Name:        util.TrimLeadingSlash(ci.Name),
			Image:       ci.Config.Image,
			ImageDigest: ci.Image,
			State:       ci.State.Status,
			StartedAt:   ci.State.StartedAt,
			Labels:      ci.Config.Labels,
		})

		if _, done := imagesSeen[ci.Image]; !done {
			imagesSeen[ci.Image] = struct{}{}
			ii, err := cli.ImageInspect(ctx, ci.Image)
			if err != nil {
				return types.WorkloadEvidence{}, fmt.Errorf("failed to inspect image %s: %w", ci.Image, err)
			}
			imageEvidence = append(imageEvidence, types.ImageEvidence{
				ID:          ii.ID,
				RepoTags:    ii.RepoTags,
				RepoDigests: ii.RepoDigests,
				Created:     ii.Created,
				Size:        ii.Size,
				Labels:      ii.Config.Labels,
			})
		}
	}

	return types.WorkloadEvidence{
		Containers: containerEvidence,
		Images:     imageEvidence,
		ReportData: challenge,
	}, nil
}
