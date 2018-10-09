package thuder

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestSimpleCopy(t *testing.T) {
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
	makeFile("a2")
	makeFile("not-a")
	makeFile("a-dir/a/a")
	makeFile("x/a")

	err := Copy(source, dest, []Filter{
		{Allow: true, NameRx: "^a"}, //copy every file that starts with a and search in every folder that starts with a
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
	if totalSize != 372 {
		t.Log(allfiles)
		t.Fatal("wrong size of total files at the end, got", totalSize)
	}
}
