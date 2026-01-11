package util

import (
	"context"
	"encoding/json"
	"fmt"
	"evidence-verifier/internal/types"
	"net/http"
)

func GetBearerToken(ctx context.Context, namespace, repo string) (token string, err error) {
	url := fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s/%s:pull", namespace, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	client := http.DefaultClient

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d fetching token", resp.StatusCode)
	}

	var tokenResponse types.TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("could not decode token response: %w", err)
	}

	return tokenResponse.Token, nil
}
