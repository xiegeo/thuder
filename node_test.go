package thuder

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// create os.FileInfo for testing
type testFileInfo struct {
	name    string
	dir     bool
	modtime time.Time
}

func (s *testFileInfo) Name() string       { return s.name }
func (s *testFileInfo) Mode() os.FileMode  { return 0777 }
func (s *testFileInfo) ModTime() time.Time { return s.modtime }
func (s *testFileInfo) IsDir() bool        { return s.dir }
func (s *testFileInfo) Sys() interface{}   { return nil }
func (s *testFileInfo) Size() int64        { return 0 }

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
			trn{"package local", cw, nil, []string{"LICENSE"}},
			trn{"file not dir", filepath.Join(cw, "LICENSE"), ErrNeedDir, nil})
	} else {
		t.Log("warning: package local files not found")
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

	expErr := func(tc trn, err error) bool {
		if err == tc.expectedError {
			return true
		}
		if tc.expectedError == os.ErrNotExist && os.IsNotExist(err) {
			return true
		}
		return false
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s with path \"%s\"", tc.testName, tc.dir), func(t *testing.T) {
			node, err := NewRootNode(tc.dir)
			if err != nil {
				if expErr(tc, err) {
					return
				}
				t.Fatal(err)
			}
			t.Logf("%s", node)

			c := NewCollection()
			err = c.Add(node)
			if err != nil {
				if expErr(tc, err) {
					return
				}
				t.Fatal(err)
			}
			for _, expect := range tc.hasFiles {
				nodes := c.Get(expect)
				if len(nodes) == 0 {
					t.Error("expected", expect)
				}
				t.Logf("found %s", nodes)
			}
			//c.PrintTo(t.Logf)
			if tc.expectedError != nil {
				t.Fatal("expected error", tc.expectedError)
			}
		})
	}
}

//fakeNodes creat fake nodes for testing with a shared FileContext
func fakeNodes(dir string, names ...string) []Node {
	ns := make([]Node, 0, len(names))
	fc := &FileContext{
		from: dir,
	}
	for _, name := range names {
		ns = append(ns, Node{
			fc: fc,
			info: &testFileInfo{
				name: name,
			},
		})
	}
	return ns
}

func TestAddNode(t *testing.T) {

}
