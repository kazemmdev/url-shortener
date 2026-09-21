package shortcode

import (
	"strings"
	"testing"
)

func TestToBase62(t *testing.T) {
	cases := []struct {
		value int64
		want  string
	}{
		{0, "0000000"},
		{61, "000000z"},
		{62, "0000010"},
	}

	for _, c := range cases {
		if got := toBase62(c.value); got != c.want {
			t.Errorf("toBase62(%d) = %q, want %q", c.value, got, c.want)
		}
	}
}

func TestGenerate(t *testing.T) {
	code, err := Generate(42)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(code) != Length {
		t.Errorf("len(code) = %d, want %d", len(code), Length)
	}
	for _, c := range code {
		if !strings.ContainsRune(alphabet, c) {
			t.Errorf("code %q contains char %q outside alphabet", code, c)
		}
	}
}

func TestGenerateVariesAcrossCalls(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 20; i++ {
		code, err := Generate(1)
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		seen[code] = true
	}
	if len(seen) < 2 {
		t.Errorf("Generate(1) produced the same code every time; shuffle looks non-random")
	}
}
