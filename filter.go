package thuder

import (
	"regexp"
	"time"
)

type Filter struct {
	Allow            bool   //allow or block if filter is matched
	Direction        string `json:",omitempty"` //(pull/push/'empty')
	FolderOnly       bool   `json:",omitempty"` //matches only folders
	FileOnly         bool   `json:",omitempty"` //matches only files
	PathRx           string `json:",omitempty"` //regexp that matches the folder path
	pathRx           *regexp.Regexp
	NameRx           string `json:",omitempty"` //regexp that matches the filename
	nameRx           *regexp.Regexp
	LastModifiedDays float32  `json:",omitempty"` //only matches within these days, 0 is unlimited
	SubFilters       []Filter `json:",omitempty"` //on match, use subfilters
}

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
