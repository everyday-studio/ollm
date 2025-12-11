package security

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

type HashParams struct {
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

var defaultParams = HashParams{
	Time:    3,
	Memory:  64 * 1024,
	Threads: 4,
	KeyLen:  32,
}

func GeneratePasswordHash(password string, p *HashParams) (string, error) {
	if len(password) == 0 {
		return "", errors.New("empty password not allowed")
	}

	if p == nil {
		p = &defaultParams
	}

	if err := validateParams(p); err != nil {
		return "", err
	}

	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, p.Time, p.Memory, p.Threads, p.KeyLen)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, p.Memory, p.Time, p.Threads, encodedSalt, encodedHash), nil
}

func validateParams(p *HashParams) error {
	// 시간 비용 검증 (OWASP 권장 최소값: 2)
	if p.Time < 2 {
		return errors.New("time cost too low (min: 2)")
	}

	// 메모리 비용 검증 (최소 32MB 권장)
	if p.Memory < 32*1024 {
		return errors.New("memory cost too low (min: 32MB)")
	}

	// 스레드 검증 (최소 1, 최대 임의 제한)
	if p.Threads < 1 {
		return errors.New("parallelism must be at least 1")
	}
	if p.Threads > 64 {
		return errors.New("parallelism exceeds maximum recommended value (max: 64)")
	}

	// 키 길이 검증 (최소 16바이트, 보통 32바이트 권장)
	if p.KeyLen < 16 {
		return errors.New("key length too short (min: 16 bytes)")
	}
	if p.KeyLen > 512 {
		return errors.New("key length exceeds maximum reasonable value (max: 512 bytes)")
	}

	// 메모리와 스레드의 비율 검증 (선택적)
	// 메모리가 충분히 크지 않으면 병렬화 이점이 줄어듦
	if p.Memory < uint32(p.Threads)*8*1024 {
		return errors.New("memory cost should be at least 8MB per thread for efficiency")
	}

	return nil
}
