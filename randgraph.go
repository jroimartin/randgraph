// Copyright (c) 2025 Roi Martin
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

// Package randgraph implements random graph generators. Graphs are
// represented as streams of edges, so full graphs are not stored in
// memory, enabling the generation of graphs of arbitrary size.
package randgraph

import (
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"strconv"
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
		lbl := ""
		if v.Label != nil {
			lbl = fmt.Sprint(v.Label)
		}
		fmt.Fprintf(w, "  %v [label=%q]\n", v.ID, lbl)
	}

	for e := range r.Edges() {
		dir := "none"
		if e.Directed {
			dir = "forward"
		}
		lbl := ""
		if e.Label != nil {
			lbl = fmt.Sprint(e.Label)
		}
		fmt.Fprintf(w, "  %v -> %v [dir=%q] [label=%q]\n", e.V0, e.V1, dir, lbl)
	}

	fmt.Fprintln(w, "}")
}

// BinomialOpts are the [Binomial] parameters.
type BinomialOpts struct {
	// N is the number of trials.
	N int

	// P is the success probability for each trial.
	P float64

	// Vertices is the number of vertices.
	Vertices int

	// Loops defines whether loops are allowed.
	Loops bool

	// Multiedges defines whether multiple edges are allowed. If
	// true, two or more edges with the same tail vertex and the
	// same head vertex are allowed.
	Multiedges bool

	// Directed defines whether the generated graphs are directed.
	Directed bool

	// Labels contains the labels used for the vertices. If the
	// number of vertices exceeds the number of available labels,
	// then duplicated labels are suffixed with the vertex number.
	Labels []string
}

// Binomial implements the [Source] interface. It generates random
// graphs in which the number of edges created per vertex follows a
// binomial distribution.
type Binomial struct {
	opts BinomialOpts
	rand *rand.Rand
}

// NewBinomial returns a new [Binomial] with the provided parameters.
func NewBinomial(opts BinomialOpts) (*Binomial, error) {
	if opts.N < 0 {
		return nil, errors.New("invalid number of trials")
	}
	if opts.P < 0 || opts.P > 1 {
		return nil, errors.New("invalid success probability")
	}
	if opts.Vertices < 0 {
		return nil, errors.New("invalid number of vertices")
	}

	b := &Binomial{
		opts: opts,
		rand: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
	}
	return b, nil
}

func (b *Binomial) Vertices() <-chan Vertex {
	ch := make(chan Vertex)
	go func() {
		for i := range b.opts.Vertices {
			ch <- Vertex{ID: i, Label: label(b.opts.Labels, i)}
		}
		close(ch)
	}()
	return ch
}

func (b *Binomial) Edges() <-chan Edge {
	ch := make(chan Edge)
	go func() {
		for tail := range b.opts.Vertices {
			var start int
			if b.opts.Loops {
				start = 0
			} else {
				if tail == b.opts.Vertices-1 {
					// No possible heads.
					break
				}
				start = tail + 1
			}

			heads := make(map[int]struct{})
			for range b.opts.N {
				if b.rand.Float64() < b.opts.P {
					head := start + b.rand.IntN(b.opts.Vertices-start)
					if !b.opts.Multiedges {
						if _, found := heads[head]; found {
							continue
						}
						heads[head] = struct{}{}
					}
					ch <- Edge{
						V0:       tail,
						V1:       head,
						Directed: b.opts.Directed,
					}
				}
			}
		}
		close(ch)
	}()
	return ch
}

func label(labels []string, id int) string {
	if len(labels) == 0 {
		return strconv.Itoa(id)
	}
	i := id % len(labels)
	if id < len(labels) {
		return labels[i]
	}
	return labels[i] + strconv.Itoa(id)
}
