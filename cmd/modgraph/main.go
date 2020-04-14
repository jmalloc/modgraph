package main

import (
	"flag"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/emicklei/dot"
	"golang.org/x/mod/modfile"
)

var (
	nextColor = 0
	colors    = []string{
		"#F0A3FF", "black", // Amethyst
		"#0075DC", "white", // Blue
		"#993F00", "white", // Caramel
		"#4C005C", "white", // Damson
		"#191919", "white", // Ebony
		"#005C31", "white", // Forest
		"#2BCE48", "black", // Green
		"#FFCC99", "black", // Honeydew
		"#808080", "white", // Iron
		"#94FFB5", "black", // Jade
		"#8F7C00", "white", // Khaki
		// "#9DCC00", "black", // Lime
		"#C20088", "white", // Mallow
		"#003380", "white", // Navy
		"#FFA405", "black", // Orpiment
		"#FFA8BB", "black", // Pink
		"#426600", "white", // Quagmire
		"#FF0010", "white", // Red
		"#5EF1F2", "black", // Sky
		"#00998F", "white", // Turquoise
		"#E0FF66", "black", // Uranium
		"#740AFF", "white", // Violet
		"#990000", "white", // Wine
		// "#FFFF80", "black", // Xanthin
		// "#FFFF00", "black", // Yellow
		"#FF5005", "black", // Zinnia
	}
)

func color() (string, string) {
	bg := colors[nextColor]
	nextColor++
	fg := colors[nextColor]
	nextColor++

	if nextColor >= len(colors) {
		nextColor = 0
	}

	return fg, bg
}

func isInModule(mod string, pkg string) bool {
	return mod == pkg || strings.HasPrefix(pkg, mod+"/")
}

func main() {
	var showInternals bool
	flag.BoolVar(&showInternals, "show-internals", false, "show usage of internal packages")
	flag.Parse()

	root := "."
	args := flag.Args()
	if len(args) > 0 {
		root = args[0]
	}

	if !filepath.IsAbs(root) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		root = path.Join(wd, root)
	}

	modpath := path.Join(root, "go.mod")
	data, err := ioutil.ReadFile(modpath)
	if err != nil {
		log.Fatal(err)
	}

	mf, err := modfile.ParseLax(modpath, data, nil)
	if err != nil {
		log.Fatal(err)
	}

	var packages []string
	imports := importMap{}

	if err := filepath.Walk(
		root,
		func(dir string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				return nil
			}

			name := info.Name()
			if strings.HasPrefix(name, ".") ||
				strings.HasPrefix(name, "_") ||
				dir == path.Join(root, "vendor") {
				return filepath.SkipDir
			}

			rel, err := filepath.Rel(root, dir)
			if err != nil {
				return err
			}

			p := mf.Module.Mod.Path
			if root != dir {
				p += "/" + rel
				rel = "./" + rel
			}

			pkg, err := build.Import(rel, root, 0)
			if _, ok := err.(*build.NoGoError); ok {
				return nil
			} else if err != nil {
				return err
			}

			for _, dep := range pkg.Imports {
				if isInModule(mf.Module.Mod.Path, dep) {
					imports.AddImport(p, dep)
				}
			}

			packages = append(packages, p)

			return nil
		},
	); err != nil {
		log.Fatal(err)
	}

	sort.Strings(packages)

	graph := dot.NewGraph(dot.Directed)
	graph.Attr("rankdir", "BT")
	graph.Attr("concentrate", "false")
	graph.Attr("splines", "true")
	graph.Attr("overlap", "false")
	graph.Attr("nodesep", "0.15")
	graph.Attr("outputorder", "edgesfirst")

	for _, p := range packages {
		n := imports.Get(p)

		if n.IsHidden(showInternals) {
			continue
		}

		node := graph.Node(p)
		node.Label(n.Label(mf.Module.Mod.Path))
		node.Attr("style", "filled")
		node.Attr("shape", "box")
		node.Attr("fontname", "Helvetica")
		node.Attr("margin", "0.15")
		node.Attr("penwidth", "2")
		node.Attr("color", "#ffffff")

		if n.IsInternal() {
			node.Attr("fontcolor", "#888888")
			node.Attr("fillcolor", "#f3f3f3")
		} else {
			fg, bg := color()
			node.Attr("fontcolor", fg)
			node.Attr("fillcolor", bg)
		}
	}

	for _, n := range imports {
		if n.IsHidden(showInternals) {
			continue
		}

		node := graph.Node(n.Path)

		for _, i := range n.Imports {
			if i.IsHidden(showInternals) {
				continue
			}

			target := graph.Node(i.Path)
			edge := graph.Edge(node, target)
			edge.Attr("penwidth", "2")
			edge.Attr("arrowsize", "0.75")

			if i.IsInternal() {
				edge.Attr("style", "dashed")
			}

			if n.IsInternal() {
				edge.Attr("color", "#dddddd")
			} else {
				edge.Attr("color", node.Value("fillcolor").(string))
			}
		}
	}

	graph.Write(os.Stdout)
}
