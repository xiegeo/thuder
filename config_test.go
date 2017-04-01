package thuder

import (
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func TestGenerateUniqueHostname(t *testing.T) {
	cmd := exec.Command("lsblk", "-O")
	cmd.Stdout = os.Stdout
	cmd.Run()

	u, err := GenerateUniqueHostname()
	if err != nil {
		t.Fatal(err)
	}
	var macpart string
	if runtime.GOOS == "windows" {
		r := regexp.MustCompile("_[0-9A-F]{4}-[0-9A-F]{4}$")
		if !r.MatchString(u) {
			t.Fatal(u, "is not of expected formate")
		}
		s := strings.Split(u, "_")
		macpart = s[len(s)-2]
	}
	t.Log(macpart)
	t.Log(u)
}
