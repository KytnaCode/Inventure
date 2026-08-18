package csrf

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
)

// generateToken generates a random, cryptographically secure, CSRF token.
func generateToken() string {
	ts := time.Now().Second()

	buf := make([]byte, 16)

	_, _ = rand.Read(buf)

	return fmt.Sprintf("%v-%v", ts, base64.RawStdEncoding.EncodeToString(buf))
}
