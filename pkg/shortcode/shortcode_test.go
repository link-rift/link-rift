package shortcode

import (
	"testing"
)

func TestGenerateDefaultLength(t *testing.T) {
	g := NewGenerator()
	code := g.Generate()
	if len(code) != DefaultLength {
		t.Errorf("expected length %d, got %d", DefaultLength, len(code))
	}
}

func TestGenerateWithLength(t *testing.T) {
	g := NewGenerator()
	for _, n := range []int{3, 5, 7, 10, 20} {
		code := g.GenerateWithLength(n)
		if len(code) != n {
			t.Errorf("expected length %d, got %d", n, len(code))
		}
	}
}

func TestGenerateUniqueness(t *testing.T) {
	g := NewGenerator()
	seen := make(map[string]bool, 10000)
	for i := 0; i < 10000; i++ {
		code := g.Generate()
		if seen[code] {
			t.Fatalf("duplicate code generated: %s (after %d iterations)", code, i)
		}
		seen[code] = true
	}
}

func TestGenerateCharset(t *testing.T) {
	g := NewGenerator()
	for i := 0; i < 100; i++ {
		code := g.Generate()
		for _, c := range code {
			found := false
			for _, valid := range charset {
				if c == valid {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("invalid character %c in code %s", c, code)
			}
		}
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	tests := []uint64{0, 1, 10, 61, 62, 100, 1000, 1000000, 18446744073709551614}
	for _, n := range tests {
		encoded := Encode(n)
		decoded, err := Decode(encoded)
		if err != nil {
			t.Fatalf("Decode(%q) error: %v", encoded, err)
		}
		if decoded != n {
			t.Errorf("round-trip failed: %d -> %q -> %d", n, encoded, decoded)
		}
	}
}

func TestDecodeInvalidCharacter(t *testing.T) {
	_, err := Decode("abc!def")
	if err != ErrInvalidCharacter {
		t.Errorf("expected ErrInvalidCharacter, got %v", err)
	}
}

func TestEncodeKnownValues(t *testing.T) {
	tests := []struct {
		input    uint64
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{61, "Z"},
		{62, "10"},
	}
	for _, tt := range tests {
		result := Encode(tt.input)
		if result != tt.expected {
			t.Errorf("Encode(%d) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
