package passhash_test

import (
	"regexp"
	"testing"

	"github.com/kytnacode/inventure/pkg/passhash"
)

var base64RegexPattern = "(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2,3})?"

var phcRegex = regexp.MustCompile(
	"\\$argon2id" + // Match argon2 variant.
		"\\$v=[0-9]+" + // Match argon2 version.
		"\\$m=[0-9]+,t=[0-9]+,p=[0-9]+" + // Match argon2 parameters.
		"\\$" + base64RegexPattern + // Base64-encoded salt.
		"\\$" + base64RegexPattern, // Base64-encoded key.
)

func TestHashShouldHavePHCFormat(t *testing.T) {
	t.Parallel()

	phcString := passhash.Hash([]byte("passowrd"))

	if !phcRegex.Match([]byte(phcString)) {
		t.Fatalf("invalid phc string: %v", phcString)
	}
}

func TestVerifyShouldVerifyCorrectHash(t *testing.T) {
	t.Parallel()

	const (
		password = "my-super-secret-password"
		phc      = "$argon2id$v=19$m=65536,t=3,p=4$uiOyD7yhvKuJwK2B+mYX9w$ZwOB3SQDZWgSI17gaRUJ5Slzr5SH8XErpN/ihlacVR8"
	)

	ok, err := passhash.Verify(phc, []byte(password))
	if err != nil {
		t.Fatalf("could not verify password: %v", err)
	}

	if !ok {
		t.Fatalf("expected verification to success")
	}
}
