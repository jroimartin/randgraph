// Copyright (c) 2025 Roi Martin
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

// Package randgraph implements random graph generators. Graphs are
// represented as streams of vertices and edges, so full graphs are
// not stored in memory, enabling the generation of graphs of
// arbitrary size.
package randgraph

import (
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
)

// A Vertex is a vertex of a graph.
type Vertex struct {
	ID    int
	Label any
}

// An Edge is an edge of a graph that connects two vertices with IDs
// V0 and V1. If directed, V0 is the tail vertex and V1 is the head
// vertex.
type Edge struct {
	ID       int
	V0, V1   int
	Directed bool
	Label    any
}

// A Source is a source of random graphs represented as streams of
// vertices and edges.
type Source interface {
	Vertices() <-chan Vertex
	Edges() <-chan Edge
}

// A RandGraph wraps a [Source] to provide higher-level functionality.
type RandGraph struct {
	src Source
}

// New returns a new [RandGraph] that uses src to generate random
// graphs.
func New(src Source) *RandGraph {
	return &RandGraph{src: src}
}

// Vertices returns a stream of random vertices.
func (r *RandGraph) Vertices() <-chan Vertex {
	return r.src.Vertices()
}

// Edges returns a stream of random edges.
func (r *RandGraph) Edges() <-chan Edge {
	return r.src.Edges()
}

// WriteDOT writes a random graph to w using the [DOT] language.
//
// [DOT]: https://graphviz.org/doc/info/lang.html
func (r *RandGraph) WriteDOT(w io.Writer) {
	fmt.Fprintln(w, "digraph {")

	for v := range r.Vertices() {
		label := ""
		if v.Label != nil {
			label = fmt.Sprint(v.Label)
		}
		fmt.Fprintf(w, "  %v [label=%q]\n", v.ID, label)
	}

	for e := range r.Edges() {
		dir := "none"
		if e.Directed {
			dir = "forward"
		}
		label := ""
		if e.Label != nil {
			label = fmt.Sprint(e.Label)
		}
		fmt.Fprintf(w, "  %v -> %v [dir=%q] [label=%q]\n", e.V0, e.V1, dir, label)
	}

	fmt.Fprintln(w, "}")
}

// Binomial implements the [Source] interface. It generates random
// graphs in which the number of edges created per vertex follows a
// binomial distribution.
type Binomial struct {
	// V is the number of vertices.
	V int

	// N is the number of edge creation trials.
	N int

	// P is the success probability for each trial.
	P float64

	// Loops defines whether loops are allowed.
	Loops bool

	// Multiedges defines whether multiple edges are allowed. If
	// true, two or more edges with the same tail vertex and the
	// same head vertex are allowed.
	Multiedges bool

	// Directed defines whether the generated graphs are directed.
	Directed bool

	// VertexLabel specifies an optional function that returns the
	// label of a vertex identified by id.
	VertexLabel func(id int) any

	// EdgeLabel specifies an optional function that returns the
	// label of an edge identified by id that connects v0 and v1.
	EdgeLabel func(id, v0, v1 int) any

	rand *rand.Rand
}

// NewBinomial returns a new [Binomial] source that generates graphs
// with v vertices. The number of edges created per vertex follows the
// binomial distribution B(n, p), where n is the number of trials and
// p the success probability for each trial.
func NewBinomial(v, n int, p float64) (*Binomial, error) {
	if v < 0 {
		return nil, errors.New("invalid number of vertices")
	}
	if n < 0 {
		return nil, errors.New("invalid number of trials")
	}
	if p < 0 || p > 1 {
		return nil, errors.New("invalid success probability")
	}

	b := &Binomial{
		V:    v,
		N:    n,
		P:    p,
		rand: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
	return b, nil
}

func (b *Binomial) Vertices() <-chan Vertex {
	ch := make(chan Vertex)
	go func() {
		for i := range b.V {
			var label any
			if b.VertexLabel != nil {
				label = b.VertexLabel(i)
			}
			ch <- Vertex{ID: i, Label: label}
		}
		close(ch)
	}()
	return ch
}

func (b *Binomial) Edges() <-chan Edge {
	ch := make(chan Edge)
	go func() {
		id := 0
		for tail := range b.V {
			var start int
			if b.Loops {
				start = 0
			} else {
				if tail == b.V-1 {
					// No possible heads.
					break
				}
				start = tail + 1
			}

			heads := make(map[int]struct{})
			for range b.N {
				if b.rand.Float64() < b.P {
					head := start + b.rand.IntN(b.V-start)
					if !b.Multiedges {
						if _, found := heads[head]; found {
							continue
						}
						heads[head] = struct{}{}
					}
					var label any
					if b.EdgeLabel != nil {
						label = b.EdgeLabel(id, tail, head)
					}
					ch <- Edge{
						ID:       id,
						V0:       tail,
						V1:       head,
						Directed: b.Directed,
						Label:    label,
					}
					id++
				}
			}
		}
		close(ch)
	}()
	return ch
}
