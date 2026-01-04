package mock

import (
	"encoding/base64"
	"os"
	"strings"
)

// IsMockMode returns true if MOCK_MODE environment variable is set to "true" or "1".
func IsMockMode() bool {
	val := strings.ToLower(os.Getenv("MOCK_MODE"))
	return val == "true" || val == "1"
}

// MockTDXQuote returns a sample TDX attestation quote in JSON format.
// The reportData field is set to the provided userData (base64-encoded challenge) for nonce-binding.
func MockTDXQuote(userData [64]byte) string {
	reportData := base64.StdEncoding.EncodeToString(userData[:])

	return `{
  "extraBytes": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
  "header": {
    "attestationKeyType": 2,
    "pceSvn": "AAA=",
    "qeSvn": "AAA=",
    "qeVendorId": "k5pyM/ecTKmUCg2zlX8GBw==",
    "teeType": 129,
    "userData": "AAAAAAAAAAAAAAAAAAAAAAAAAAA=",
    "version": 4
  },
  "signedData": {
    "certificationData": {
      "certificateDataType": 6,
      "qeReportCertificationData": {
        "pckCertificateChainData": {
          "certificateDataType": 5,
          "pckCertChain": "LS0tLS1NT0NLX0NFUlRJRklDQVRFX0NIQUlOLS0tLS0K",
          "size": 100
        },
        "qeAuthData": {
          "data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
          "parsedDataSize": 32
        },
        "qeReport": {
          "attributes": "FQAAAAAAAADnAAAAAAAAAA==",
          "cpuSvn": "CAj/GwT/AAYAAAAAAAAAAA==",
          "isvProdId": 2,
          "isvSvn": 6,
          "mrEnclave": "5aOntdgwwpU7mFNMbFmjo0/cNOkz9/WJjwqFzwiEa8o=",
          "mrSigner": "3J4qfG+UjxdHTjSn/EPtAw98FWPxur3fY0DILg5UqMU=",
          "reportData": "/7S0wmcRqFU+sr7qsb8ZPzl4Vs7Yzy4Q8kPTwUmFDpIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
          "reserved1": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA==",
          "reserved2": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
          "reserved3": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
          "reserved4": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
        },
        "qeReportSignature": "6+wY07cReoHeYX5McWsANL0HVvc39t7mV6Gs6ClUR295V8F7Spe7TavzQsmnejlCNdEob5abOj8m+IWa0fWVnw=="
      },
      "size": 200
    },
    "ecdsaAttestationKey": "rKB2N4wx/HSJCDDMcHlNtfpkq4N0cNGbn/jZaC9Qr0lCkZUiOOowljdxqpI7NJbZKOP1dlA3YKZik7lY8r6BMQ==",
    "signature": "aFMYAr4N2hRd4yA1kBS+cvoPZQa4l0zE6Htj1M6M+CPU6Djv84ZQ1lpsTp/O1vY+bGe3LuXTbIsjPAcRZEGl6A=="
  },
  "signedDataSize": 200,
  "tdQuoteBody": {
    "mrConfigId": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "mrOwner": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "mrOwnerConfig": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "mrSeam": "v7NgrI5iM6G8oUM8r3OC2VwWW0p3+wC/FDXloI8wDN/q1e5oRhr9m2xyjc51NGAt",
    "mrSignerSeam": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
    "mrTd": "pYROiIl7cMMYvvkp7039bHMExSxLycPzkTLw/czs8+tbq3ARDuQqElCaMcA3KIaU",
    "reportData": "` + reportData + `",
    "rtmrs": [
      "CnNnuY2/Pq2D/If2tRcOfF7q+vD85kVXKJIXizJK89O67NAiCVPTumlePUO/V2cd",
      "B9LExWMygRmB9mzTgZP+YgM7+9Roili72kwde5lJJDWC5mU0lpMr04KFFqu49eFi",
      "8fYsVRVmoobfTAeaafi7X1856Thjjbzi9uPZcWOkpojIYCIShokNeVNZTRbE9Dai",
      "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
    ],
    "seamAttributes": "AAAAAAAAAAA=",
    "tdAttributes": "AAAAEAAAAAA=",
    "teeTcbSvn": "CAEIAAAAAAAAAAAAAAAAAA==",
    "xfam": "5wAGAAAAAAA="
  }
}`
}

// MockWorkloadEvidence returns sample workload evidence with the challenge bound to reportData.
func MockWorkloadEvidence(challenge string) (containers []map[string]any, images []map[string]any) {
	containers = []map[string]any{
		{
			"id":          "90f62b5f43c6a1c06e6df91d97476fbf6d624e30663c4fabf3400ef40886cbe7",
			"name":        "llm-core",
			"image":       "sanctuairy/llm-core:latest",
			"imageDigest": "sha256:435faa6db70075a575dd54e2a1e76cee14bd53fb67be4fdfa3736d879b9f1ccb",
			"state":       "running",
			"startedAt":   "2025-12-08T13:41:20.710779927Z",
			"labels": map[string]string{
				"org.opencontainers.image.ref.name": "ubuntu",
				"org.opencontainers.image.version":  "24.04",
			},
		},
	}
	images = []map[string]any{
		{
			"id":       "sha256:435faa6db70075a575dd54e2a1e76cee14bd53fb67be4fdfa3736d879b9f1ccb",
			"repoTags": []string{"sanctuairy/llm-core:latest"},
			"repoDigests": []string{
				"sanctuairy/llm-core@sha256:1fc562c7ec052c28bd87c1e0d788a2f542fd177559e78a5db9cfd258db0b3d51",
			},
			"created": "2025-11-13T13:35:24.498087954Z",
			"size":    int64(3653158191),
			"labels": map[string]string{
				"org.opencontainers.image.ref.name": "ubuntu",
				"org.opencontainers.image.version":  "24.04",
			},
		},
	}
	return containers, images
}

// MockInfrastructureSummary returns sample infrastructure summary for mock mode.
func MockInfrastructureSummary() map[string]string {
	return map[string]string{
		"provider":    "Google Cloud Platform (Mock)",
		"instanceId":  "9214714451018122242",
		"name":        "cvm-1-mock",
		"zone":        "europe-west4-a",
		"machineType": "c3-standard-4",
		"status":      "RUNNING",
		"projectId":   "cvm-icdcs-mock",
	}
}
