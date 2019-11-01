package main

import (
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
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

func isInternal(pkg string) bool {
	return strings.Contains(pkg, "/internal/") ||
		strings.HasSuffix(pkg, "/internal")
}

func isParent(parent, child string) bool {
	return strings.HasPrefix(child, parent+"/")
}

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
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

	graph := dot.NewGraph(dot.Directed)
	graph.Attr("rankdir", "BT")
	graph.Attr("concentrate", "false")
	graph.Attr("splines", "true")
	graph.Attr("overlap", "false")
	graph.Attr("nodesep", "0.15")
	graph.Attr("outputorder", "edgesfirst")

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

			disp := strings.ReplaceAll(rel, "/", "\n")
			qual := mf.Module.Mod.Path
			if root != dir {
				qual += "/" + rel
				rel = "./" + rel
			}

			pkg, err := build.Import(rel, root, 0)
			if _, ok := err.(*build.NoGoError); ok {
				return nil
			} else if err != nil {
				return err
			}

			if disp == "." {
				disp = "(" + pkg.Name + ")"
			}

			node := graph.Node(qual)
			node.Label(disp)
			node.Attr("style", "filled")
			node.Attr("shape", "box")
			node.Attr("fontname", "Helvetica")
			node.Attr("margin", "0.15")
			node.Attr("penwidth", "2")
			node.Attr("color", "#ffffff")
			node.Attr("rank", strconv.Itoa(len(qual)))

			edgeColor := "#eeeeee"
			if isInternal(qual) {
				node.Attr("fontcolor", "#888888")
				node.Attr("fillcolor", "#f3f3f3")
			} else {
				fg, bg := color()
				edgeColor = bg
				node.Attr("fontcolor", fg)
				node.Attr("fillcolor", bg)
			}

			for _, p := range pkg.Imports {
				if isInModule(mf.Module.Mod.Path, p) {
					targeted := graph.Node(p)

					edge := graph.Edge(node, targeted)
					edge.Attr("penwidth", "2")
					edge.Attr("color", edgeColor)
					edge.Attr("arrowsize", "0.75")

					if isInternal(p) {
						edge.Attr("style", "dashed")
					}

					if isParent(p, qual) {
						edge.Attr("constraint", "false")

						con := graph.Edge(targeted, node)
						con.Attr("constraint", "true")
						con.Attr("style", "invis")
					}
				}
			}

			return nil
		},
	); err != nil {
		log.Fatal(err)
	}

	graph.Write(os.Stdout)
}
