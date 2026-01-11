package manifest

import (
	"fmt"
	"evidence-verifier/internal/types"
)

func VerifyQuoteValues(manifestUrl string, evidenceQuoteBody types.TdQuoteBody) error {
	baselineManifest, err := FetchManifest(manifestUrl)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest from %s: %w", manifestUrl, err)
	}

	if baselineManifest.MrTd != evidenceQuoteBody.MrTd {
		return fmt.Errorf("manifest MR-TD mismatch: expected %s, got %s", baselineManifest.MrTd, evidenceQuoteBody.MrTd)
	}

	if baselineManifest.MrSeam != evidenceQuoteBody.MrSeam {
		return fmt.Errorf("manifest MR-SEAM mismatch: expected %s, got %s", baselineManifest.MrSeam, evidenceQuoteBody.MrSeam)
	}

	if baselineManifest.TdAttributes != evidenceQuoteBody.TdAttributes {
		return fmt.Errorf("manifest TD attributes mismatch: expected %s, got %s", baselineManifest.TdAttributes, evidenceQuoteBody.TdAttributes)
	}

	if baselineManifest.Xfam != evidenceQuoteBody.Xfam {
		return fmt.Errorf("manifest Xfam mismatch: expected %s, got %s", baselineManifest.Xfam, evidenceQuoteBody.Xfam)
	}

	if baselineManifest.TeeTcbSvn != evidenceQuoteBody.TeeTcbSvn {
		return fmt.Errorf("manifest TEE TCB SVN mismatch: expected %s, got %s", baselineManifest.TeeTcbSvn, evidenceQuoteBody.TeeTcbSvn)
	}

	return nil
}
