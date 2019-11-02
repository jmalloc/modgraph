package main

import "strings"

type node struct {
	Path       string
	Imports    map[string]*node
	ImportedBy map[string]*node
}

func (n *node) Label(mod string) string {
	if n.Path == mod {
		chunks := strings.Split(mod, "/")
		return "(" + chunks[len(chunks)-1] + ")"
	}

	rel := strings.TrimPrefix(n.Path, mod+"/")
	return strings.ReplaceAll(rel, "/", "\n")
}

func (n *node) IsInternal() bool {
	return strings.Contains(n.Path, "/internal/") || strings.HasSuffix(n.Path, "/internal")
}

func (n *node) IsHidden() bool {
	if !n.IsInternal() {
		return false
	}

	for _, i := range n.ImportedBy {
		if !i.IsHidden() {
			return false
		}
	}

	return true
}

type importMap map[string]*node

func (m importMap) Get(p string) *node {
	n, ok := m[p]
	if !ok {
		n = &node{
			Path:       p,
			Imports:    map[string]*node{},
			ImportedBy: map[string]*node{},
		}
		m[p] = n
	}

	return n
}

func (m importMap) AddImport(from, to string) {
	f := m.Get(from)
	t := m.Get(to)

	f.Imports[to] = t
	t.ImportedBy[from] = f
}
