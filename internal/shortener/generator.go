package shortener

import (
	"context"
	"crypto/rand"
	"math/big"
)

const Base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type CodeGenerator interface {
	Generate(ctx context.Context) (string, error)
}

type RandomGenerator struct {
	length int
}

func NewRandomGenerator(length int) RandomGenerator {
	if length <= 0 {
		length = 8
	}
	return RandomGenerator{length: length}
}

func (g RandomGenerator) Generate(ctx context.Context) (string, error) {
	code := make([]byte, g.length)
	max := big.NewInt(int64(len(Base62Alphabet)))

	for i := range code {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		code[i] = Base62Alphabet[n.Int64()]
	}

	return string(code), nil
}
