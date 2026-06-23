package shortener

import (
	"context"
	"strings"
	"testing"
)

func TestRandomGeneratorProducesBase62Code(t *testing.T) {
	t.Parallel()

	generator := NewRandomGenerator(12)

	code, err := generator.Generate(context.Background())
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if len(code) != 12 {
		t.Fatalf("len(code) = %d, want 12", len(code))
	}
	for _, char := range code {
		if !strings.ContainsRune(Base62Alphabet, char) {
			t.Fatalf("code contains non-Base62 character %q", char)
		}
	}
}
