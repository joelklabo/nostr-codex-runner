package mailgun

import "testing"

func TestVerifySignature(t *testing.T) {
	signingKey := "test-key"
	timestamp := "1700000000"
	token := "abcdef"

	// Precomputed with HMAC-SHA256(signing_key, timestamp + token)
	expectedSig := "b1ba822482ac05022a79329bb189ad0bba81a33a164bcaf7e97e2c886ff159c4"

	if !verifySignature(timestamp, token, expectedSig, signingKey) {
		t.Fatalf("expected signature to verify")
	}

	if verifySignature(timestamp, token, "deadbeef", signingKey) {
		t.Fatalf("expected signature mismatch to fail")
	}
}
