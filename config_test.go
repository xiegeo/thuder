package thuder

import (
	"testing"
)

func TestGenerateUniqueHostname(t *testing.T) {
	u, err := GenerateUniqueHostname()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(u)
}
