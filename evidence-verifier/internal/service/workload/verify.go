package workload

import (
	"fmt"
	"evidence-verifier/internal/service/challenge"
	"evidence-verifier/internal/types"
)

func Verify(issuedChallenge string, workloadEvidence *types.WorkloadEvidence, referenceImage *types.ReferenceImage, workloadReferenceMetadata *types.WorkloadReferenceMetadata) (types.VerificationReport, error) {
	imageUri := fmt.Sprintf("%s/%s:%s", referenceImage.Namespace, referenceImage.Repository, referenceImage.Tag)

	workloadToVerify, err := findWorkload(workloadEvidence.ContainerEvidence, imageUri)
	if err != nil {
		return types.VerificationReport{}, err
	}

	// 1. Verify the challenge issued by the Relying Party.
	err = challenge.Verify(issuedChallenge, workloadEvidence.ReportData)
	if err != nil {
		return types.VerificationReport{
			IsVerified: false,
			Message:    err.Error(),
		}, nil
	}

	// 2. Verify the workload evidence against the Reference Value Provider's golden image values.
	goldenWorkloadImageDigest := workloadReferenceMetadata.Manifest.Config.Digest

	if workloadToVerify.ImageDigest != goldenWorkloadImageDigest {
		return types.VerificationReport{
			IsVerified: false,
			Message:    fmt.Sprintf("workload image digest mismatch: expected %s, got %s", goldenWorkloadImageDigest, workloadToVerify.ImageDigest),
		}, nil
	}

	return types.VerificationReport{
		IsVerified:          true,
		Message:             fmt.Sprintf("successfully verified image for workload '%s'", workloadToVerify.Name),
		GoldenImageDigest:   goldenWorkloadImageDigest,
		ProvidedImageDigest: workloadToVerify.ImageDigest,
		ReferenceImage:      imageUri,
	}, nil
}

func findWorkload(workloads []types.ContainerEvidence, imageUri string) (types.ContainerEvidence, error) {
	for _, workload := range workloads {
		if workload.Image == imageUri {
			return workload, nil
		}
	}

	return types.ContainerEvidence{}, fmt.Errorf("no container found for image %s", imageUri)
}
