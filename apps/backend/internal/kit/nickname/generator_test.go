package nickname

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

func TestGenerate(t *testing.T) {
	nickname, err := Generate()
	assert.NoError(t, err)
	assert.NotEmpty(t, nickname)

	// Since we are using Korean adjectives and nouns, length should be greater than a few characters
	assert.GreaterOrEqual(t, utf8.RuneCountInString(nickname), 2)
}

func TestGenerateFallback(t *testing.T) {
	nickname := GenerateFallback()
	assert.Equal(t, "이름없는올름", nickname)
}
