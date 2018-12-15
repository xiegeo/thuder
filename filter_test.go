package thuder

import (
	"fmt"
	"testing"
	"time"
)

func TestFilter(t *testing.T) {
	n := &fakeNodes("test", "abc")[0]
	f := Filter{}
	now := time.Now()
	t.Run(fmt.Sprintf("Direction"), func(t *testing.T) {
		f.Direction = "foobar"
		expErr(t, f.Prepare())
		f.Direction = "push"
		noErr(t, f.Prepare())
		m, a := f.Match(n, true, now)
		expect(t, m, a, true, false)
		f.Allow = true
		m, a = f.Match(n, true, now)
		expect(t, m, a, true, true)
		m, a = f.Match(n, false, now)
		expect(t, m, a, false, false)
		f.Direction = "pull"
		m, a = f.Match(n, false, now)
		expect(t, m, a, true, true)
		m, a = f.Match(n, true, now)
		expect(t, m, a, false, false)
		f.Direction = ""
		m, a = f.Match(n, true, now)
		expect(t, m, a, true, true)
		m, a = f.Match(n, false, now)
		expect(t, m, a, true, true)
	})
	matches := func(t *testing.T, em bool) {
		m, a := f.Match(n, false, now)
		expect(t, m, a, em, em)
	}
	t.Run(fmt.Sprintf("FolderOnly"), func(t *testing.T) {
		matches(t, true)
		f.FolderOnly = true
		matches(t, false)
		n.info.(*testFileInfo).dir = true
		matches(t, true)
	})
	t.Run(fmt.Sprintf("FileOnly"), func(t *testing.T) {
		f.FileOnly = true
		expErr(t, f.Prepare())
		f.FolderOnly = false
		noErr(t, f.Prepare())
		matches(t, false)
		n.info.(*testFileInfo).dir = false
		matches(t, true)
	})
	t.Run(fmt.Sprintf("PathRx"), func(t *testing.T) {
		f.PathRx = "\\"
		expErr(t, f.Prepare())
		f.PathRx = "testtest"
		noErr(t, f.Prepare())
		matches(t, false)
		f.PathRx = "test"
		noErr(t, f.Prepare())
		matches(t, true)
	})
	t.Run(fmt.Sprintf("NameRx"), func(t *testing.T) {
		f.NameRx = "\\"
		expErr(t, f.Prepare())
		f.NameRx = "x"
		noErr(t, f.Prepare())
		matches(t, false)
		f.NameRx = "a"
		noErr(t, f.Prepare())
		matches(t, true)
	})
	t.Run(fmt.Sprintf("LastModifiedDays"), func(t *testing.T) {
		n.info.(*testFileInfo).modtime = now.Add(-30 * time.Hour)
		f.LastModifiedDays = -1
		expErr(t, f.Prepare())
		f.LastModifiedDays = 1
		noErr(t, f.Prepare())
		matches(t, false)
		f.LastModifiedDays = 2
		noErr(t, f.Prepare())
		matches(t, true)
	})
	t.Run(fmt.Sprintf("SubFilters"), func(t *testing.T) {
		f2 := f
		f.SubFilters = []Filter{f2}
		matches(t, true)
		f.Allow = false
		matches(t, true)
		f.SubFilters[0].LastModifiedDays = -1
		expErr(t, f.Prepare())
		f.SubFilters[0].LastModifiedDays = 1
		noErr(t, f.Prepare())
		m, a := f.Match(n, false, now)
		expect(t, m, a, true, false)
		f.SubFilters = append(f.SubFilters, f2)
		matches(t, true)
		f.LastModifiedDays = 1
		matches(t, false)
	})

}

func expect(t *testing.T, m, a, em, ea bool) {
	if m != em || a != ea {
		t.Errorf("got (%v,%v) expected (%v,%v)", m, a, em, ea)
	}
}

func expErr(t *testing.T, err error) {
	if err == nil {
		t.Error("error expected")
	} else {
		t.Log("expected error: ", err)
	}
}

func noErr(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}
