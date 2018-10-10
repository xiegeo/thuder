package thuder

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestCopy(t *testing.T) {
	fs2 := fs
	defer func() { fs = fs2 }()
	afs := afero.NewMemMapFs()
	fs = wrapAfero(afs)

	source := "/source/"

	dest := "/dest/"

	fm := os.FileMode(0777)

	fs.MkdirAll(source, fm)
	fs.MkdirAll(dest, fm)

	makeFile := func(name string) {
		err := afero.WriteFile(afs, source+name, []byte(name), fm)
		if err != nil {
			t.Fatal(err)
		}
	}

	makeFile("a")
	makeFile("a1")
	makeFile("not-a")
	makeFile("a-dir/a/a")
	makeFile("x/a")

	type cs struct {
		rx   string
		size int64
	}

	testCases := []cs{
		{rx: "^a", size: 368},           //copy everything but not-a and x/a
		{rx: "^.$", size: 368 + 42 + 3}, //update a, add x/a
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("copy %v", tc.rx), func(t *testing.T) {
			err := Copy(source, dest, []Filter{
				{Allow: true, NameRx: tc.rx},
			})
			if err != nil {
				t.Fatal(err)
			}

			var totalSize int64
			allfiles := bytes.NewBuffer(nil)
			afero.Walk(afs, "/", func(path string, info os.FileInfo, err error) error {
				fmt.Fprintln(allfiles, path, info.Size())
				totalSize += info.Size()
				return nil
			})
			if totalSize != tc.size {
				t.Log(allfiles)
				t.Fatal("wrong size of total files at the end, got", totalSize)
			}
		})
	}

}
