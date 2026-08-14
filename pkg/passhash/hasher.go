// Package passhash implements secure password hashing using argon2id.
package passhash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"

	"golang.org/x/crypto/argon2"
)

const argon2idVariant = "argon2id"

// ErrInvalidFormat is returned if hash doesn't conform to PHC format, or is hashed using a
// different algorithm or argon2id version.
var ErrInvalidFormat = errors.New("invalid hash format or version")

// Recommended parameters.
var defaultConfig = &config{
	saltLength: 16,
	keyLength:  32,
	time:       3,
	memory:     64 * 1024,
	threads:    4,
}

// config is argon2id configuration.
type config struct {
	saltLength uint32
	keyLength  uint32
	time       uint32
	memory     uint32
	threads    uint8
}

// genSalt generates a new cryptographically secure hash, can panic if underlying OS random
// number generator fails.
func genSalt(saltLen uint32) []byte {
	salt := make([]byte, saltLen)

	_, _ = rand.Read(salt)

	return salt
}

// Hash hashes a clear text password and return the password hash in PHC format.
func Hash(clearTextPassword []byte) (phcString string) {
	salt := genSalt(defaultConfig.saltLength)

	conf := defaultConfig

	rawHash := hash(clearTextPassword, salt, conf)

	return formatHash(rawHash, salt, conf)
}

// hash a password with the given salt and parameters.
func hash(clearTextPassword, salt []byte, conf *config) (rawHash []byte) {
	time := conf.time
	memory := conf.memory
	threads := conf.threads
	keyLength := conf.keyLength

	hashed := argon2.IDKey(clearTextPassword, salt, time, memory, threads, keyLength)

	return hashed
}

// Verify takes a PHC-formatted string and cryptographically securely compares if a clear
// text password matches the given hash.
func Verify(phcString string, clearTextPassword []byte) (ok bool, err error) {
	rawHash, salt, conf, err := scanHash(phcString)
	if err != nil {
		return false, err
	}

	computedHash := hash(clearTextPassword, salt, conf)

	return subtle.ConstantTimeCompare(rawHash, computedHash) == 1, nil
}

// formatHash PHC-formats a hash.
func formatHash(hash, salt []byte, conf *config) string {
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)

	return fmt.Sprintf(
		"$%v$v=%v$m=%v,t=%v,p=%v$%v$%v",
		argon2idVariant,
		argon2.Version,
		conf.time,
		conf.memory,
		conf.threads,
		saltEncoded,
		hashEncoded,
	)
}

// scanHash get hash key, salt, and parameters from a PHC string.
func scanHash(phcString string) (hash, salt []byte, conf *config, err error) {
	var version int

	conf = new(config)

	const componentsNum = 5

	parts := strings.Split(phcString, "$")

	components := parts[1:]

	if len(components) != componentsNum {
		return nil, nil, nil, fmt.Errorf(
			"phc string must be composed of 5 dollar sign separed components: %w",
			ErrInvalidFormat,
		)
	}

	variant := components[0]

	if variant != argon2idVariant {
		return nil, nil, nil, fmt.Errorf("hash must use argon2id algorithm")
	}

	_, err = fmt.Sscanf(components[1], "v=%v", &version)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid version format: %w", err)
	}

	if version != argon2.Version {
		return nil, nil, nil, fmt.Errorf("unsupported argon2 version")
	}

	_, err = fmt.Sscanf(
		components[2],
		"m=%v,t=%v,p=%v",
		&conf.memory,
		&conf.time,
		&conf.threads,
	)
	if err != nil {
		return nil, nil, nil, errors.Join(ErrInvalidFormat, err)
	}

	salt, err = base64.RawStdEncoding.DecodeString(components[3])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid base64 salt: %w", err)
	}

	hash, err = base64.RawStdEncoding.DecodeString(components[4])
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid base64 hash: %w", err)
	}

	saltLen := len(salt)

	if saltLen > math.MaxUint32 {
		return nil, nil, nil, fmt.Errorf(
			"salt length must fit in a unsigned 32 bit integer: %w",
			ErrInvalidFormat,
		)
	}

	conf.saltLength = uint32(saltLen)

	keyLen := len(hash)

	if keyLen > math.MaxUint32 {
		return nil, nil, nil, fmt.Errorf(
			"key length must fit in a unsigned 32 bit integer: %w",
			ErrInvalidFormat,
		)
	}

	conf.keyLength = uint32(keyLen)

	return hash, salt, conf, nil
}
