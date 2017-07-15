package thuder

import (
	"fmt"
	"regexp"
	"time"
)

type Filter struct {
	Allow            bool   //allow or block if filter is matched
	Direction        string `json:",omitempty"` //(pull/push/"")
	FolderOnly       bool   `json:",omitempty"` //matches only folders
	FileOnly         bool   `json:",omitempty"` //matches only files
	PathRx           string `json:",omitempty"` //regexp that matches the folder path
	pathRx           *regexp.Regexp
	NameRx           string `json:",omitempty"` //regexp that matches the filename
	nameRx           *regexp.Regexp
	LastModifiedDays float32  `json:",omitempty"` //only matches within these days, 0 is unlimited
	SubFilters       []Filter `json:",omitempty"` //on match, use subfilters
}

//Prepare prepares a Filter for usage, and return any errors in filter defination
func (f *Filter) Prepare() error {
	if len(f.Direction) > 0 && f.Direction != "pull" && f.Direction != "push" {
		return fmt.Errorf("unknown Direction %v should be (pull/push/\"\")", f.Direction)
	}
	if f.FolderOnly && f.FileOnly {
		return fmt.Errorf("both FolderOnly and FileOnly")
	}
	var err error
	f.pathRx, err = regexp.Compile(f.PathRx)
	if err != nil {
		return fmt.Errorf("PathRx %v", err)
	}
	f.nameRx, err = regexp.Compile(f.NameRx)
	if err != nil {
		return fmt.Errorf("NameRx %v", err)
	}
	if f.LastModifiedDays < 0 {
		return fmt.Errorf("LastModifiedDays (%v) should be positive", f.LastModifiedDays)
	}
	return PrepareFilters(f.SubFilters)
}

func PrepareFilters(fs []Filter) error {
	for i := 0; i < len(fs); i++ {
		err := fs[i].Prepare()
		if err != nil {
			return fmt.Errorf("sub(%v) %v", i, err)
		}
	}
	return nil
}

//Match tests if Node matches filter, and if matched, is it allowed or blocked
func (f *Filter) Match(n *Node, push bool, now time.Time) (match bool, allow bool) {
	if len(f.Direction) > 1 {
		if push && f.Direction != "push" {
			return false, false
		}
		if !push && f.Direction != "pull" {
			return false, false
		}
	}
	if n.IsDir() {
		if f.FileOnly {
			return false, false
		}
	} else if f.FolderOnly {
		return false, false
	}
	if f.pathRx != nil && !f.pathRx.MatchString(n.fc.from) {
		return false, false
	}
	if f.nameRx != nil && !f.nameRx.MatchString(n.info.Name()) {
		return false, false
	}
	if f.LastModifiedDays != 0 {
		duration := time.Duration(f.LastModifiedDays * float32(time.Hour*24))
		if n.ModTime().Add(duration).Before(now) {
			return false, false
		}
	}
	m, a := MatchFilters(f.SubFilters, n, push, now)

	if m {
		return m, a
	}

	return true, f.Allow
}

func MatchFilters(fs []Filter, n *Node, push bool, now time.Time) (match bool, allow bool) {
	for _, sub := range fs {
		m, a := sub.Match(n, push, now)
		if m {
			return m, a
		}
	}
	return false, false
}
