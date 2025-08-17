// Copyright (c) 2025 Roi Martin
// Use of this source code is governed by the MIT license that can be
// found in the LICENSE file.

package randgraph

import (
	"bytes"
	"math/rand/v2"
	"regexp"
	"slices"
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
			V0:    0,
			V1:    1,
			Label: "e0",
		},
		{
			V0:       1,
			V1:       2,
			Directed: true,
		},
		{
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
			V0:    0,
			V1:    1,
			Label: "e0",
		},
		{
			V0:       1,
			V1:       2,
			Directed: true,
		},
		{
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
		opts       BinomialOpts
		wantNilErr bool
	}{
		{
			name:       "zero",
			opts:       BinomialOpts{},
			wantNilErr: true,
		},
		{
			name: "Vertices < 0",
			opts: BinomialOpts{
				Vertices: -1,
			},
			wantNilErr: false,
		},
		{
			name: "N < 0",
			opts: BinomialOpts{
				N: -1,
			},
			wantNilErr: false,
		},
		{
			name: "P < 0",
			opts: BinomialOpts{
				P: -0.1,
			},
			wantNilErr: false,
		},
		{
			name: "P > 1",
			opts: BinomialOpts{
				P: 1.1,
			},
			wantNilErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := NewBinomial(tt.opts)
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
		name string
		opts BinomialOpts
		want []Edge
	}{
		{
			name: "1 vertex with loops and multiedges",
			opts: BinomialOpts{
				Vertices:   1,
				N:          2,
				P:          1,
				Loops:      true,
				Multiedges: true,
			},
			want: []Edge{
				{V0: 0, V1: 0},
				{V0: 0, V1: 0},
			},
		},
		{
			name: "1 vertex with multiedges",
			opts: BinomialOpts{
				Vertices:   1,
				N:          2,
				P:          1,
				Loops:      false,
				Multiedges: true,
			},
			want: []Edge{},
		},
		{
			name: "1 vertex",
			opts: BinomialOpts{
				Vertices:   1,
				N:          2,
				P:          1,
				Loops:      false,
				Multiedges: false,
			},
			want: []Edge{},
		},
		{
			name: "2 vertices with multiedges",
			opts: BinomialOpts{
				Vertices:   2,
				N:          2,
				P:          1,
				Loops:      false,
				Multiedges: true,
			},
			want: []Edge{
				{V0: 0, V1: 1},
				{V0: 0, V1: 1},
			},
		},
		{
			name: "2 vertices",
			opts: BinomialOpts{
				Vertices:   2,
				N:          2,
				P:          1,
				Loops:      false,
				Multiedges: false,
			},
			want: []Edge{
				{V0: 0, V1: 1},
			},
		},
		{
			name: "directed",
			opts: BinomialOpts{
				Vertices:   2,
				N:          1,
				P:          1,
				Loops:      false,
				Multiedges: false,
				Directed:   true,
			},
			want: []Edge{
				{V0: 0, V1: 1, Directed: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Binomial{
				opts: tt.opts,
				rand: testRand(),
			}

			var got []Edge
			for e := range b.Edges() {
				got = append(got, e)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("unexpected edges: got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestBinomial_edgeless(t *testing.T) {
	const numVertices = 5

	b := &Binomial{
		opts: BinomialOpts{
			Vertices: numVertices,
			N:        0,
			P:        0,
		},
		rand: testRand(),
	}

	var gotVs []Vertex
	for v := range b.Vertices() {
		gotVs = append(gotVs, v)
	}
	if n := len(gotVs); n != numVertices {
		t.Errorf("unexpected number of vertices: got: %v, want: %v", n, numVertices)
	}

	var gotEs []Edge
	for e := range b.Edges() {
		gotEs = append(gotEs, e)
	}
	if n := len(gotEs); n != 0 {
		t.Errorf("expected number of edges: got: %v, want: 0", n)
	}
}

func TestBinomial_order_zero(t *testing.T) {
	b := &Binomial{
		opts: BinomialOpts{
			Vertices: 0,
			N:        1,
			P:        1,
		},
		rand: testRand(),
	}

	var gotVs []Vertex
	for v := range b.Vertices() {
		gotVs = append(gotVs, v)
	}
	if n := len(gotVs); n != 0 {
		t.Errorf("unexpected number of vertices: got: %v, want: 0", n)
	}

	var gotEs []Edge
	for e := range b.Edges() {
		gotEs = append(gotEs, e)
	}
	if n := len(gotEs); n != 0 {
		t.Errorf("expected number of edges: got: %v, want: 0", n)
	}
}

func TestLabel(t *testing.T) {
	tests := []struct {
		labels []string
		id     int
		want   string
	}{
		{
			labels: []string{"A", "B"},
			id:     0,
			want:   "A",
		},
		{
			labels: []string{"A", "B"},
			id:     1,
			want:   "B",
		},
		{
			labels: []string{"A", "B"},
			id:     2,
			want:   "A2",
		},
		{
			labels: []string{"A", "B"},
			id:     5,
			want:   "B5",
		},
		{
			labels: nil,
			id:     2,
			want:   "2",
		},
		{
			labels: []string{},
			id:     5,
			want:   "5",
		},
	}
	for _, tt := range tests {
		got := label(tt.labels, tt.id)
		if got != tt.want {
			t.Errorf("unexpected label: got: %q, want: %q", got, tt.want)
		}
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
