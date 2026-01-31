package shortcode

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
)

const (
	charset       = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base          = uint64(len(charset))
	DefaultLength = 7
)

var ErrInvalidCharacter = errors.New("invalid base62 character")

type Generator interface {
	Generate() string
	GenerateWithLength(n int) string
}

type generator struct{}

func NewGenerator() Generator {
	return &generator{}
}

func (g *generator) Generate() string {
	return g.GenerateWithLength(DefaultLength)
}

func (g *generator) GenerateWithLength(n int) string {
	b := make([]byte, n)
	max := big.NewInt(int64(len(charset)))
	for i := range b {
		idx, err := rand.Int(rand.Reader, max)
		if err != nil {
			panic("crypto/rand failed: " + err.Error())
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b)
}

func Encode(n uint64) string {
	if n == 0 {
		return string(charset[0])
	}

	var sb strings.Builder
	for n > 0 {
		sb.WriteByte(charset[n%base])
		n /= base
	}

	// Reverse
	result := []byte(sb.String())
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

func Decode(s string) (uint64, error) {
	var n uint64
	for _, c := range s {
		idx := strings.IndexRune(charset, c)
		if idx < 0 {
			return 0, ErrInvalidCharacter
		}
		n = n*base + uint64(idx)
	}
	return n, nil
}
