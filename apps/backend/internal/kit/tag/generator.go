package tag

import (
	"crypto/rand"
	"math/big"
)

const charset = "123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const tagLength = 5

// Generate generates a random 5-character string using characters 1-9 and A-Z.
func Generate() (string, error) {
	tag := make([]byte, tagLength)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < tagLength; i++ {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		tag[i] = charset[idx.Int64()]
	}

	return string(tag), nil
}
