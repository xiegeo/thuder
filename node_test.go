package thuder

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

type trn struct {
	testName      string
	dir           string
	expectedError error
	hasFiles      []string
}

func TestRootNode(t *testing.T) {

	dne, _ := filepath.Abs("does_not_exist.file")

	testCases := []trn{
		{"relative", "abc", ErrBadPath, nil},
		{"does not exist", dne, os.ErrNotExist, nil}, //ErrNotExist makes it test for os.IsNotExist
	}

	cw, _ := filepath.Abs(".")
	if filepath.Base(cw) == "thuder" {
		testCases = append(testCases,
			trn{"package local", cw, nil, []string{"LICENSE"}})
	}

	var osCases []trn
	if filepath.Separator == '/' {
		osCases = []trn{
			{"root", "/", nil, []string{"dev"}},
		}
	} else {
		osCases = []trn{
			{"root", `C:\`, nil, []string{"Program Files"}},
		}
	}

	testCases = append(testCases, osCases...)

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s with path \"%s\"", tc.testName, tc.dir), func(t *testing.T) {
			node, err := NewRootNode(tc.dir)
			if err != nil {
				if err == tc.expectedError {
					return
				}
				if tc.expectedError == os.ErrNotExist && os.IsNotExist(err) {
					return
				}
				t.Fatal(err)
			}
			t.Logf("%s", node)

			c := NewCollection()
			err = c.Add(node)
			if err != nil {
				t.Fatal(err)
			}
			for _, expect := range tc.hasFiles {
				n, ok := c.nodes[expect]
				if !ok {
					t.Error("expected", expect)
				}
				t.Logf("found %s", &n)
			}
			//c.PrintTo(t.Logf)
			if tc.expectedError != nil {
				t.Fatal("expected error", tc.expectedError)
			}
		})
	}

}
