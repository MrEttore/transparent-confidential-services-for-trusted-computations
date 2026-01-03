# Testing TLS-Bound Attestation (Mock Mode)

This guide describes how to test the **TLS-Bound Attestation** flow end-to-end using Docker Compose and Mock Mode.

## Prerequisites

- **Docker** and **Docker Compose**
- **curl** (for making HTTP requests)
- **jq** (recommended, for parsing JSON)

## 1. Start Services

The `docker-compose.yml` is already configured with `MOCK_MODE=true` and an **Nginx Gateway**.

```bash
docker-compose up --build
```
*Wait until all services are running. The Evidence Provider is accessible via the Nginx gateway at `https://localhost:8443`.*

## 2. Manual Testing Steps

You can run these commands one by one to simulate the Relying Application's flow.

### Step A: Generate a Challenge
The client (Relying Party) generates a random nonce.

```bash
export CHALLENGE_B64="MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMw=="
echo "Challenge: $CHALLENGE_B64"
```

### Step B: Get Quote from Evidence Provider (via Nginx Gateway)
Request a TDX quote through the secure gateway. We use `-k` (insecure) because the certificate is self-signed.

```bash
RESPONSE=$(curl -k -s -X POST https://localhost:8443/evidence/tdx-quote \
  -H "Content-Type: application/json" \
  -d "{\"challenge\": \"$CHALLENGE_B64\"}")

echo $RESPONSE | jq .
```

### Step C: Extract Data for Verification
Extract the `quote` and the `tlsCertificateFingerprint` returned by the Provider.

```bash
export QUOTE=$(echo $RESPONSE | jq -r '.data.quote')
export FINGERPRINT=$(echo $RESPONSE | jq -r '.data.tlsCertificateFingerprint')

echo "Fingerprint: $FINGERPRINT"
```

### Step D: Verify with Evidence Verifier
Send the extracted data to the Verifier (port 8081). The Verifier checks that the Quote's `UserData` matches `SHA256(Challenge || Fingerprint)`.

```bash
curl -s -X POST http://localhost:8081/verify/tdx-quote \
  -H "Content-Type: application/json" \
  -d "{
    \"issuedChallenge\": \"$CHALLENGE_B64\",
    \"baselineManifestUrl\": \"http://example.com/manifest\",
    \"tlsCertificateFingerprint\": \"$FINGERPRINT\",
    \"quote\": $QUOTE
  }" | jq .
```

**Expected Result:**
```json
{
  "status": "success",
  "data": {
    "isVerified": true,
    "message": "tdx quote verified (mock mode - hardware verification skipped)"
  }
}
```

---

## 3. Automated Test Script (One-Liner)

Copy and run this entire block to perform the full test automatically (using the HTTPS gateway):

```bash
# 1. Challenge (64 bytes base64)
CHALLENGE="MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDEyMw=="

# 2. Get Quote (via HTTPS Gateway :8443)
echo "Fetching Quote via HTTPS Gateway..."
RESP=$(curl -k -s -X POST https://localhost:8443/evidence/tdx-quote -d "{\"challenge\": \"$CHALLENGE\"}")
QUOTE=$(echo $RESP | jq -r '.data.quote')
FINGERPRINT=$(echo $RESP | jq -r '.data.tlsCertificateFingerprint')

echo "Fingerprint: $FINGERPRINT"

# 3. Verify (via Verifier :8081)
echo "Verifying..."
curl -s -X POST http://localhost:8081/verify/tdx-quote \
  -H "Content-Type: application/json" \
  -d "{
    \"issuedChallenge\": \"$CHALLENGE\",
    \"baselineManifestUrl\": \"http://example.com/manifest\",
    \"tlsCertificateFingerprint\": \"$FINGERPRINT\",
    \"quote\": $QUOTE
  }" | jq .
```
