package workload

import (
	"context"
	"encoding/json"
	"evidence-verifier/internal/service/mock"
	"evidence-verifier/internal/types"
	"evidence-verifier/internal/util"
	"fmt"
	"net/http"
)

// FetchImageTagMetadata fetches tag‐level info for a given namespace/repo/tag.
//
//	namespace: e.g., "sanctuairy"
//	repo:      e.g., "llm-core"
//	tag:       e.g., "latest"
func fetchImageTagMetadata(
	ctx context.Context,
	namespace, repo, tag string,
) (types.TagMetadata, error) {

	url := fmt.Sprintf(
		"https://hub.docker.com/v2/repositories/%s/%s/tags/%s",
		namespace, repo, tag,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return types.TagMetadata{}, fmt.Errorf("failed to create a request for Docker Hub’s Hub API: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "evidence-verifier-service/1.0")

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return types.TagMetadata{}, fmt.Errorf("request for Docker Hub’s Hub API failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.TagMetadata{}, fmt.Errorf("unexpected status %d from Docker Hub’s Hub API", resp.StatusCode)
	}

	var metadata types.TagMetadata
	if err := json.NewDecoder(resp.Body).Decode(&metadata); err != nil {
		return types.TagMetadata{}, fmt.Errorf("failed to decode JSON (fetchImageTagMetadata): %w", err)
	}

	return metadata, nil
}

// fetchImageManifestV2 retrieves the full v2 manifest for namespace/repo/tag.
//
//	namespace: e.g. "sanctuairy"
//	repo:      e.g. "llm-core"
//	reference: can be a tag ("latest") or a digest ("sha256:…")
func fetchImageManifestV2(
	ctx context.Context,
	namespace, repo, reference string,
) (types.ManifestV2, error) {
	token, err := util.GetBearerToken(ctx, namespace, repo)
	if err != nil {
		return types.ManifestV2{}, fmt.Errorf("failed to fetch Bearer token: %w", err)
	}

	url := fmt.Sprintf(
		"https://registry-1.docker.io/v2/%s/%s/manifests/%s",
		namespace, repo, reference,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return types.ManifestV2{}, fmt.Errorf("failed to create a request for Docker Hub’s Registry HTTP API V2: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return types.ManifestV2{}, fmt.Errorf("manifest request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return types.ManifestV2{}, fmt.Errorf("unexpected status %d fetching manifest", resp.StatusCode)
	}

	var manifest types.ManifestV2
	if err := json.NewDecoder(resp.Body).Decode(&manifest); err != nil {
		return types.ManifestV2{}, fmt.Errorf("failed to decode JSON manifest: %w", err)
	}

	return manifest, nil
}

// FetchWorkloadReferenceMetadata retrieves both the tag‐level metadata and the full v2 manifest
// for a given Docker image. It calls Docker Hub’s Hub API to fetch tag information (name, digest,
// etc.) and then the Registry HTTP API V2 to fetch the manifest (including config.digest).
//
// Parameters:
//   - ctx:       the context for request cancellation and timeouts
//   - namespace: the Docker namespace (e.g., "sanctuairy")
//   - repo:      the repository name (e.g., "llm-core")
//   - tag:       the image reference (e.g., "latest" or a specific digest)
//
// Returns:
//   - *types.WorkloadReferenceMetadata: a struct containing both TagMetadata and ManifestV2
//   - error: if any HTTP call or JSON decoding fails, an error is returned
func FetchWorkloadReferenceMetadata(
	ctx context.Context,
	referenceImage *types.ReferenceImage,
) (types.WorkloadReferenceMetadata, error) {
	if mock.IsMockMode() {
		// Return mock metadata matching the provider's mock values
		// Digest: sha256:435faa6db70075a575dd54e2a1e76cee14bd53fb67be4fdfa3736d879b9f1ccb
		return types.WorkloadReferenceMetadata{
			Metadata: types.TagMetadata{
				Name: "latest",
				Images: []types.TagImage{
					{Digest: "sha256:435faa6db70075a575dd54e2a1e76cee14bd53fb67be4fdfa3736d879b9f1ccb"},
				},
			},
			Manifest: types.ManifestV2{
				Config: struct {
					MediaType string `json:"mediaType"`
					Size      int64  `json:"size"`
					Digest    string `json:"digest"`
				}{
					Digest: "sha256:435faa6db70075a575dd54e2a1e76cee14bd53fb67be4fdfa3736d879b9f1ccb",
				},
			},
		}, nil
	}

	metadata, err := fetchImageTagMetadata(ctx, referenceImage.Namespace, referenceImage.Repository, referenceImage.Tag)
	if err != nil {
		return types.WorkloadReferenceMetadata{}, err
	}

	manifest, err := fetchImageManifestV2(ctx, referenceImage.Namespace, referenceImage.Repository, referenceImage.Tag)
	if err != nil {
		return types.WorkloadReferenceMetadata{}, err
	}

	workloadReferenceMetadata := types.WorkloadReferenceMetadata{
		Metadata: metadata,
		Manifest: manifest,
	}

	return workloadReferenceMetadata, nil

}
