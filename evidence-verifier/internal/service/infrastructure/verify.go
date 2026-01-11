package infrastructure

import (
	"evidence-verifier/internal/service/challenge"
	"evidence-verifier/internal/service/manifest"
	"evidence-verifier/internal/service/mock"
	"evidence-verifier/internal/types"
	"fmt"
)

func Verify(issuedChallenge, manifestUrl string, infrastructureEvidence types.InfrastructureEvidence) (types.VerificationReport, error) {

	// 1. Verify the challenge issued by the Relying Party.
	if err := challenge.Verify(issuedChallenge, infrastructureEvidence.ReportData); err != nil {
		return types.VerificationReport{
			IsVerified: false,
			Message:    err.Error(),
		}, nil
	}

	// 2. Verify the infrastructure evidence against the Reference Value Provider's manifest.
	// In Mock Mode, we skip fetching the remote manifest to avoid failures on dead links.
	if !mock.IsMockMode() {
		if err := manifest.VerifyInfrastructureValues(manifestUrl, infrastructureEvidence); err != nil {
			return types.VerificationReport{
				IsVerified: false,
				Message:    fmt.Sprintf("manifest verification failed: %v", err),
			}, nil
		}
	}

	return types.VerificationReport{
		IsVerified:                    true,
		Message:                       "infrastructure successfully verified",
		VerifiedBootDiskSourceImage:   infrastructureEvidence.Disk.SourceImage,
		VerifiedBootDiskSourceImageID: infrastructureEvidence.Disk.SourceImageId,
	}, nil
}
