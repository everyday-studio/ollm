package nickname

import (
	"crypto/rand"
	"math/big"
)

var adjectives = []string{
	"영원한", "잊혀진", "심연의", "잠든",
	"지혜로운", "빛나는", "아스라한",
	"고대의", "신비로운", "황혼의",
}

var nouns = []string{
	"올름", "메아리", "감시자", "프롬프트",
	"탐구자", "동굴", "숲",
	"기억", "환상", "이방인",
}

// Generate creates a random nickname by combining an adjective and a noun.
// The concept is based on the Ollm universe.
func Generate() (string, error) {
	adjIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(adjectives))))
	if err != nil {
		return "", err
	}

	nounIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(nouns))))
	if err != nil {
		return "", err
	}

	return adjectives[adjIndex.Int64()] + nouns[nounIndex.Int64()], nil
}

// GenerateFallback returns a fallback nickname if Generate fails.
func GenerateFallback() string {
	return "이름없는올름"
}
