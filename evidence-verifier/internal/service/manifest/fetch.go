package manifest

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"evidence-verifier/internal/types"
	"io"
	"net/http"
	"strings"
)

// FetchManifest fetches a JSONC (comments-allowed) manifest,
// strips out any // comments, then decodes into the Go struct.
func FetchManifest(url string) (types.BaselineManifest, error) {
	resp, err := http.Get(url)
	if err != nil {
		return types.BaselineManifest{}, fmt.Errorf("failed to GET manifest at %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return types.BaselineManifest{}, fmt.Errorf(
			"unexpected HTTP status %d from %s: %s",
			resp.StatusCode, url, resp.Status,
		)
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return types.BaselineManifest{}, fmt.Errorf("failed to read manifest body: %w", err)
	}

	var cleaned bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			continue
		}
		cleaned.WriteString(line)
		cleaned.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return types.BaselineManifest{}, fmt.Errorf("error scanning manifest body: %w", err)
	}

	var manifest types.BaselineManifest
	if err := json.Unmarshal(cleaned.Bytes(), &manifest); err != nil {
		return types.BaselineManifest{}, fmt.Errorf("failed to decode JSON manifest: %w", err)
	}

	return manifest, nil
}
