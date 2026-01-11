package attestationquote

import (
	"fmt"
	"log"

	"evidence-provider/internal/service/mock"
	"github.com/google/go-tdx-guest/client"
	"github.com/google/go-tdx-guest/proto/tdx"
	"google.golang.org/protobuf/encoding/protojson"
)

// FetchEvidence retrieves an Intel TDX attestation quote based on a client-defined challenge (userData).
// It interacts with the TDX Quote Provider to fetch the attestation quote, converts it to JSON format,
// and returns it as a string.
//
// If MOCK_MODE environment variable is set, returns sample mock data for local development.
//
// Parameters:
//   - userData: A fixed-length array ([64]byte) containing the user challenge to be included in the attestation quote (field reportData).
//
// Returns:
//   - string: The attestation quote in JSON format.
//   - error: An error if the quote retrieval or conversion fails.
func FetchEvidence(userData [64]byte) (string, error) {
	if mock.IsMockMode() {
		log.Println("Mock mode enabled: returning sample TDX quote")
		return mock.MockTDXQuote(userData), nil
	}

	quoteProvider, err := client.GetQuoteProvider()
	if err != nil {
		return "", fmt.Errorf("failed to get Quote Provider: %v", err)
	}

	protocolBufferQuote, err := client.GetQuote(quoteProvider, userData)
	if err != nil {
		return "", fmt.Errorf("failed to get attestation quote: %v", err)
	}

	quote, ok := protocolBufferQuote.(*tdx.QuoteV4)
	if !ok {
		return "", fmt.Errorf("failed to assert protocolBufferQuote as *tdx.QuoteV4")
	}

	quoteJSON, err := protojson.Marshal(quote)
	if err != nil {
		return "", fmt.Errorf("failed to convert quote to JSON: %v", err)
	}

	return string(quoteJSON), nil
}
