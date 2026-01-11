package manifest

import (
	"fmt"
	"evidence-verifier/internal/types"
)

func VerifyInfrastructureValues(manifestUrl string, infrastructureEvidence types.InfrastructureEvidence) error {
	baselineManifest, err := FetchManifest(manifestUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest from %s: %w", manifestUrl, err)
	}

	if baselineManifest.SourceImage != infrastructureEvidence.Disk.SourceImage {
		return fmt.Errorf("manifest source image mismatch: expected %s, got %s", baselineManifest.SourceImage, infrastructureEvidence.Disk.SourceImage)
	}

	if baselineManifest.SourceImageID != infrastructureEvidence.Disk.SourceImageId {
		return fmt.Errorf("manifest source image ID mismatch: expected %s, got %s", baselineManifest.SourceImageID, infrastructureEvidence.Disk.SourceImageId)
	}

	return nil
}
