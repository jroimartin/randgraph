// Copyright (c) 2025 Roi Martin
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package randgraph

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"regexp"
	"slices"
	"strconv"
	"testing"
)

func TestRandGraph_Vertices(t *testing.T) {
	want := []Vertex{
		{
			ID:    0,
			Label: "v0",
		},
		{
			ID: 1,
		},
		{
			ID:    2,
			Label: "v2",
		},
	}

	r := New(newTestSource(want, nil))

	var got []Vertex
	for v := range r.Vertices() {
		got = append(got, v)
	}

	if !slices.Equal(got, want) {
		t.Errorf("unexpected vertices: got: %v, want: %v", got, want)
	}
}

func TestRandGraph_Edges(t *testing.T) {
	want := []Edge{
		{
			ID:    0,
			V0:    0,
			V1:    1,
			Label: "e0",
		},
		{
			ID:       1,
			V0:       1,
			V1:       2,
			Directed: true,
		},
		{
			ID:       2,
			V0:       2,
			V1:       0,
			Directed: true,
			Label:    "e2",
		},
	}

	r := New(newTestSource(nil, want))

	var got []Edge
	for e := range r.Edges() {
		got = append(got, e)
	}

	if !slices.Equal(got, want) {
		t.Errorf("unexpected edges: got: %v, want: %v", got, want)
	}
}

var validDOT = regexp.MustCompile(`(?m)^digraph {\n(  \d+ \[label="[^"]*"\]\n)+(  \d -> \d \[dir="(forward|none)"\] \[label="[^"]*"\]\n)+}$`)

func TestRandGraph_WriteDOT(t *testing.T) {
	vertices := []Vertex{
		{
			ID:    0,
			Label: "v0",
		},
		{
			ID: 1,
		},
		{
			ID:    2,
			Label: "v2",
		},
	}
	edges := []Edge{
		{
			ID:    0,
			V0:    0,
			V1:    1,
			Label: "e0",
		},
		{
			ID:       1,
			V0:       1,
			V1:       2,
			Directed: true,
		},
		{
			ID:       2,
			V0:       2,
			V1:       0,
			Directed: true,
			Label:    "e2",
		},
	}

	r := New(newTestSource(vertices, edges))
	buf := &bytes.Buffer{}
	r.WriteDOT(buf)
	out := buf.String()
	if !validDOT.MatchString(out) {
		t.Errorf("malformed output:\n%v", out)
	}
}

func TestNewBinomial(t *testing.T) {
	tests := []struct {
		name       string
		v          int
		n          int
		p          float64
		wantNilErr bool
	}{
		{
			name:       "zero",
			wantNilErr: true,
		},
		{
			name:       "v < 0",
			v:          -1,
			wantNilErr: false,
		},
		{
			name:       "n < 0",
			n:          -1,
			wantNilErr: false,
		},
		{
			name:       "p < 0",
			p:          -0.1,
			wantNilErr: false,
		},
		{
			name:       "p > 1",
			p:          1.1,
			wantNilErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBinomial(tt.v, tt.n, tt.p)
			if (err == nil) != tt.wantNilErr {
				t.Errorf("unexpected error: %v", err)
			}
			if (b == nil) != !tt.wantNilErr {
				t.Errorf("unexpected value: %v", b)
			}
		})
	}
}

func TestBinomial(t *testing.T) {
	tests := []struct {
		name        string
		v           int
		n           int
		p           float64
		loops       bool
		multiedges  bool
		directed    bool
		vertexLabel func(id int) any
		edgeLabel   func(id, v0, v1 int) any
		wantVs      []Vertex
		wantEs      []Edge
	}{
		{
			name:       "1 vertex with loops and multiedges",
			v:          1,
			n:          5,
			p:          1,
			loops:      true,
			multiedges: true,
			wantVs: []Vertex{
				{ID: 0},
			},
			wantEs: []Edge{
				{ID: 0, V0: 0, V1: 0},
				{ID: 1, V0: 0, V1: 0},
				{ID: 2, V0: 0, V1: 0},
				{ID: 3, V0: 0, V1: 0},
				{ID: 4, V0: 0, V1: 0},
			},
		},
		{
			name:       "1 vertex with multiedges",
			v:          1,
			n:          5,
			p:          1,
			multiedges: true,
			wantVs: []Vertex{
				{ID: 0},
			},
			wantEs: []Edge{},
		},
		{
			name: "1 vertex",
			v:    1,
			n:    5,
			p:    1,
			wantVs: []Vertex{
				{ID: 0},
			},
			wantEs: []Edge{},
		},
		{
			name:       "2 vertices with multiedges",
			v:          2,
			n:          5,
			p:          1,
			multiedges: true,
			wantVs: []Vertex{
				{ID: 0},
				{ID: 1},
			},
			wantEs: []Edge{
				{ID: 0, V0: 0, V1: 1},
				{ID: 1, V0: 0, V1: 1},
				{ID: 2, V0: 0, V1: 1},
				{ID: 3, V0: 0, V1: 1},
				{ID: 4, V0: 0, V1: 1},
			},
		},
		{
			name: "2 vertices",
			v:    2,
			n:    5,
			p:    1,
			wantVs: []Vertex{
				{ID: 0},
				{ID: 1},
			},
			wantEs: []Edge{
				{ID: 0, V0: 0, V1: 1},
			},
		},
		{
			name: "edgeless with n=0",
			v:    5,
			n:    0,
			p:    1,
			wantVs: []Vertex{
				{ID: 0},
				{ID: 1},
				{ID: 2},
				{ID: 3},
				{ID: 4},
			},
			wantEs: []Edge{},
		},
		{
			name: "edgeless with p=0",
			v:    5,
			n:    5,
			p:    0,
			wantVs: []Vertex{
				{ID: 0},
				{ID: 1},
				{ID: 2},
				{ID: 3},
				{ID: 4},
			},
			wantEs: []Edge{},
		},
		{
			name:   "order zero",
			v:      0,
			n:      5,
			p:      1,
			wantVs: []Vertex{},
			wantEs: []Edge{},
		},
		{
			name:     "directed",
			v:        2,
			n:        1,
			p:        1,
			directed: true,
			wantVs: []Vertex{
				{ID: 0},
				{ID: 1},
			},
			wantEs: []Edge{
				{ID: 0, V0: 0, V1: 1, Directed: true},
			},
		},
		{
			name: "vertex label",
			v:    2,
			n:    1,
			p:    1,
			vertexLabel: func(id int) any {
				return strconv.Itoa(id)
			},
			wantVs: []Vertex{
				{ID: 0, Label: "0"},
				{ID: 1, Label: "1"},
			},
			wantEs: []Edge{
				{ID: 0, V0: 0, V1: 1},
			},
		},
		{
			name: "edge label",
			v:    2,
			n:    1,
			p:    1,
			edgeLabel: func(id, v0, v1 int) any {
				return fmt.Sprintf("%v-%v-%v", id, v0, v1)
			},
			wantVs: []Vertex{
				{ID: 0},
				{ID: 1},
			},
			wantEs: []Edge{
				{ID: 0, V0: 0, V1: 1, Label: "0-0-1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBinomial(tt.v, tt.n, tt.p)
			if err != nil {
				t.Fatal(err)
			}
			b.Loops = tt.loops
			b.Multiedges = tt.multiedges
			b.Directed = tt.directed
			b.VertexLabel = tt.vertexLabel
			b.EdgeLabel = tt.edgeLabel
			b.rand = testRand()

			var gotVs []Vertex
			for v := range b.Vertices() {
				gotVs = append(gotVs, v)
			}
			if !slices.Equal(gotVs, tt.wantVs) {
				t.Errorf("unexpected vertices: got: %v, want: %v", gotVs, tt.wantVs)
			}

			var gotEs []Edge
			for e := range b.Edges() {
				gotEs = append(gotEs, e)
			}
			if !slices.Equal(gotEs, tt.wantEs) {
				t.Errorf("unexpected edges: got: %v, want: %v", gotEs, tt.wantEs)
			}
		})
	}
}

type testSource struct {
	vertices []Vertex
	edges    []Edge
}

func newTestSource(vertices []Vertex, edges []Edge) testSource {
	return testSource{
		vertices: vertices,
		edges:    edges,
	}
}

func (src testSource) Vertices() <-chan Vertex {
	ch := make(chan Vertex)
	go func() {
		for _, v := range src.vertices {
			ch <- v
		}
		close(ch)
	}()
	return ch

}
func (src testSource) Edges() <-chan Edge {
	ch := make(chan Edge)
	go func() {
		for _, e := range src.edges {
			ch <- e
		}
		close(ch)
	}()
	return ch
}

func testRand() *rand.Rand {
	return rand.New(rand.NewPCG(1, 2))
}
