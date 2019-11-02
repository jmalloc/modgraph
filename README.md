# modgraph

Produce graphviz diagrams of import relationships between packages within a
single Go module.

## Installation

```
GO111MODULE=off go get -u github.com/jmalloc/modgraph/...
```

## Usage

```
cd /path/to/module
modgraph | dot -Tpng -o /tmp/graph.png
```
