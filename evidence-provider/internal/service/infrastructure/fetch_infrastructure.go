package infrastructure

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
	"evidence-provider/internal/service/mock"
	"evidence-provider/internal/types"
	"evidence-provider/internal/util"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// FetchEvidence retrieves infrastructure evidence based on the detected cloud provider.
// It gathers metadata, instance details, and disk information for supported providers.
//
// If MOCK_MODE environment variable is set, returns sample mock data for local development.
func FetchEvidence(ctx context.Context, challenge string) (types.InfrastructureEvidence, error) {
	if mock.IsMockMode() {
		log.Println("Mock mode enabled: returning sample infrastructure evidence")

		return types.InfrastructureEvidence{
			Summary: &types.InfrastructureSummary{
				Provider:    "Google Cloud Platform (Mock)",
				InstanceID:  "9214714451018122242",
				Name:        "cvm-1-mock",
				Zone:        "europe-west4-a",
				MachineType: "c3-standard-4",
				Status:      "RUNNING",
				ProjectID:   "cvm-icdcs-mock",
			},
			Instance: &compute.Instance{
				Name:              "cvm-1-mock",
				Id:                9214714451018122242,
				CpuPlatform:       "Intel Sapphire Rapids",
				Status:            "RUNNING",
				CreationTimestamp: "2025-12-07T09:35:41.982-08:00",
				ConfidentialInstanceConfig: &compute.ConfidentialInstanceConfig{
					ConfidentialInstanceType: "TDX",
				},
			},
			Disk: &compute.Disk{
				Name:        "cvm-1-mock",
				Id:          9057367825912955906,
				SizeGb:      30,
				Status:      "READY",
				SourceImage: "https://www.googleapis.com/compute/v1/projects/cvm-icdcs/global/images/golden-reference-tee",
			},
			ReportData: challenge,
		}, nil
	}

	summary := &types.InfrastructureSummary{}
	var instance *compute.Instance
	var disk *compute.Disk
	var err error

	switch detectCloudProvider() {
	case "GCP":
		summary.Provider = "Google Cloud Platform"
		err = util.FillGCP(ctx, summary)
		if err != nil {
			return types.InfrastructureEvidence{}, fmt.Errorf("could not fill GCP metadata: %w", err)
		}

		instance, err = fetchGCPInstance(ctx, summary.ProjectID, summary.Zone, summary.Name)
		if err != nil {
			return types.InfrastructureEvidence{}, fmt.Errorf("could not fetch GCP instance: %w", err)
		}

		disk, err = fetchGCPInstanceBootDisk(ctx, summary.ProjectID, summary.Zone, instance.Name)
		if err != nil {
			return types.InfrastructureEvidence{}, fmt.Errorf("could not fetch GCP disk: %w", err)
		}

	case "AWS":
		summary.Provider = "Amazon Web Services"
		// TODO: For future work
		// ...
		return types.InfrastructureEvidence{}, fmt.Errorf("provider not supported yet")

	case "Azure":
		summary.Provider = "Microsoft Azure"
		// TODO: For future work
		// ...
		return types.InfrastructureEvidence{}, fmt.Errorf("provider not supported yet")

	default:
		return types.InfrastructureEvidence{}, fmt.Errorf("unknown cloud provider")
	}

	return types.InfrastructureEvidence{
		Summary:    summary,
		Instance:   instance,
		Disk:       disk,
		ReportData: challenge,
	}, nil

}

// detectCloudProvider determines the cloud provider where the code is running.
//
// Returns:
//   - string: The name of the detected cloud provider ("GCP", "AWS", "Azure", or "Unknown").
func detectCloudProvider() string {

	// Report whether this process is running on the GCP.
	if metadata.OnGCE() {
		return "GCP"
	}

	// Report whether this process is running on AWS.
	if reachable("http://169.254.169.254/latest/meta-data/") {
		return "AWS"
	}

	// Report whether this process is running on Azure.
	if reachableWithHeader("http://169.254.169.254/metadata/instance?api-version=2021-02-01",
		map[string]string{"Metadata": "true"}) {
		return "Azure"
	}

	return "Unknown"

}

// fetchGCPInstance retrieves the metadata for a Google Compute Engine instance.
func fetchGCPInstance(ctx context.Context, projectID, zone, instanceName string) (*compute.Instance, error) {
	computeClient, err := compute.NewService(ctx, option.WithScopes(compute.ComputeReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("compute.NewService: %w", err)
	}

	inst, err := computeClient.Instances.Get(projectID, zone, instanceName).Do()
	if err != nil {
		return nil, fmt.Errorf(
			"Instances.Get(%q,%q,%q): %w",
			projectID, zone, instanceName, err,
		)
	}

	return inst, nil
}

// fetchGCPInstanceBootDisk retrieves the metadata for a Google Compute Engine boot disk.
//
// Note: This function assumes the VM has exactly one boot disk.
func fetchGCPInstanceBootDisk(ctx context.Context, projectID, zone, diskName string) (*compute.Disk, error) {
	computeClient, err := compute.NewService(ctx, option.WithScopes(compute.ComputeReadonlyScope))
	if err != nil {
		return nil, fmt.Errorf("compute.NewService: %w", err)
	}

	disk, err := computeClient.Disks.Get(projectID, zone, diskName).Do()
	if err != nil {
		return nil, fmt.Errorf(
			"Disks.Get(%q, %q, %q): %w", projectID, zone, diskName, err,
		)
	}

	return disk, nil
}

// reachable checks if a given URL is reachable.
func reachable(url string) bool {
	client := http.Client{Timeout: 200 * time.Millisecond}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// reachableWithHeader checks if a given URL is reachable with specific headers.
func reachableWithHeader(url string, headers map[string]string) bool {
	client := http.Client{Timeout: 200 * time.Millisecond}
	req, _ := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
