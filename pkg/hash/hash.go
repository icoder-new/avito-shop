package hash

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

type Config struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

type Hasher struct {
	config Config
}

func NewHasher(cfg Config) *Hasher {
	if cfg.Time == 0 {
		cfg.Time = 1
	}
	if cfg.Memory == 0 {
		cfg.Memory = 64 * 1024
	}
	if cfg.Threads == 0 {
		cfg.Threads = 4
	}
	if cfg.KeyLen == 0 {
		cfg.KeyLen = 32
	}

	return &Hasher{
		config: cfg,
	}
}

func (h *Hasher) Hash(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.config.Time,
		h.config.Memory,
		h.config.Threads,
		h.config.KeyLen,
	)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.config.Memory,
		h.config.Time,
		h.config.Threads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func (h *Hasher) Verify(password, hash string) (bool, error) {
	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid hash format")
	}

	if parts[1] != "argon2id" {
		return false, errors.New("invalid hash algorithm")
	}

	if parts[2] != "v=19" {
		return false, errors.New("invalid argon2 version")
	}

	var memory uint32
	var time uint32
	var threads uint8

	_, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return false, fmt.Errorf("failed to parse hash parameters: %w", err)
	}

	if memory != h.config.Memory || time != h.config.Time || threads != h.config.Threads {
		return false, errors.New("hash parameters don't match current config")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("failed to decode hash: %w", err)
	}

	if uint32(len(decodedHash)) != h.config.KeyLen {
		return false, errors.New("invalid key length")
	}

	newHash := argon2.IDKey(
		[]byte(password),
		salt,
		time,
		memory,
		threads,
		h.config.KeyLen,
	)

	return subtle.ConstantTimeCompare(newHash, decodedHash) == 1, nil
}
