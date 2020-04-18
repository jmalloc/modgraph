package main

import (
	"fmt"
	"path"
	"strings"
)

type hideSet map[string]struct{}

func (h hideSet) String() string {
	return fmt.Sprintf("%#v", h)
}

func (h hideSet) Set(value string) error {
	h[value] = struct{}{}
	return nil
}

type visibility struct {
	ModulePath    string
	ShowInternal  bool
	HideShallow   hideSet
	HideRecursive hideSet
}

func (v *visibility) IsHidden(n *node) bool {
	rel := "."
	if n.Path != v.ModulePath {
		rel = strings.TrimSuffix(strings.TrimPrefix(n.Path, v.ModulePath+"/"), "/")
	}

	if _, ok := v.HideShallow[rel]; ok {
		return true
	}

	for rel != "." {
		if _, ok := v.HideRecursive[rel]; ok {
			return true
		}

		rel = path.Dir(rel)
	}

	if !v.ShowInternal {
		return n.IsInternal()
	}

	if !n.IsInternal() {
		return false
	}

	for _, i := range n.ImportedBy {
		if !v.IsHidden(i) {
			return false
		}
	}

	return true
}
