package mailgun

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// verifySignature checks Mailgun webhook signatures.
// See https://documentation.mailgun.com/docs/mailgun/api-reference/webhooks/signed-webhooks/
func verifySignature(timestamp, token, signature, signingKey string) bool {
	mac := hmac.New(sha256.New, []byte(signingKey))
	mac.Write([]byte(timestamp))
	mac.Write([]byte(token))
	expected := mac.Sum(nil)
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	return hmac.Equal(expected, sigBytes)
}
