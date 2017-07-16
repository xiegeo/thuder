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
			node, err := NewRootNode(tc.dir, false)
			if err != nil {
				if expErr(tc, err) {
					return
				}
				t.Fatal(err)
			}
			t.Logf("%s", node)

			c := NewCollection(acceptAll)
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
	dir, _ := filepath.Abs(".")
	as := []string{"aa", "AA", "aA", "Aa"} //listed out of order
	ns := fakeNodes(dir, as...)
	var out []Node
	last := func() Node {
		return out[len(out)-1]
	}

	for _, n := range ns {
		out = addNode(out, n)
	}

	if len(out) != len(as) {
		t.Errorf("not all in the same dir are added (%v/%v)", len(out), len(as))
	}

	if last().info.Name() != "AA" {
		t.Errorf("AA should be ordered last, but got %v", last().info.Name())
	}

	//t.Log(out)
}

func TestEqualFileModTime(t *testing.T) {
	testCases := []struct {
		a, b float32
		same bool
	}{
		{1, 1, true},
		{1, 1.125, true},
		{1, 1.725, true},
		{1, 1.875, true},
		{1, 2, true},
		{2, 3, true},
		{0, 2, false},
		{0.0001, 2, true},
		{0.0001, 1.5, false},
		{0.0001, 1, true},
		{0.0001, 0.6, false},
		{0.0001, 0.5, true},
		{0, 10, false},
	}
	toTime := func(s float32) time.Time {
		return time.Unix(0, 0).Add(time.Duration(s * float32(time.Second)))
	}
	for i, tc := range testCases {
		a := toTime(tc.a)
		b := toTime(tc.b)
		if EqualFileModTime(a, b) != tc.same {
			t.Errorf("EqualFileModTime (%v) for %v and %v expected %v, but got %v",
				i, a, b, tc.same, EqualFileModTime(a, b))
		}
		if EqualFileModTime(b, a) != tc.same {
			t.Errorf("EqualFileModTime (%v') for %v and %v expected %v, but got %v",
				i, b, a, tc.same, EqualFileModTime(b, a))
		}
	}

}
