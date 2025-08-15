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

func TestRandGraph_Graph(t *testing.T) {
	want := []Edge{
		{
			V0: "v0",
			V1: "v1",
		},
		{
			V0:       "v1",
			V1:       "v2",
			Directed: true,
		},
		{
			V0: "v2",
		},
	}

	r := New(newTestSource(want))

	var got []Edge
	for edge := range r.Graph() {
		got = append(got, edge)
	}

	if !slices.Equal(got, want) {
		t.Errorf("unexpected edges: got: %v want: %v", got, want)
	}
}

var validText = regexp.MustCompile(`(?m)^(\w+( -[->] \w+)?\n)+$`)

func TestRandGraph_Write(t *testing.T) {
	edges := []Edge{
		{
			V0:       "v0",
			V1:       "v1",
			Directed: true,
		},
		{
			V0: "v1",
			V1: "v2",
		},
		{
			V0: "v2",
		},
	}

	r := New(newTestSource(edges))
	buf := &bytes.Buffer{}
	r.Write(buf)
	out := buf.String()
	if !validText.MatchString(out) {
		t.Errorf("malformed output:\n%v", out)
	}
}

var validDOT = regexp.MustCompile(`(?m)^digraph {\n(  "\w+"( -> "\w+"( \[dir = none\])?)?\n)+}$`)

func TestRandGraph_WriteDOT(t *testing.T) {
	edges := []Edge{
		{
			V0:       "v0",
			V1:       "v1",
			Directed: true,
		},
		{
			V0: "v1",
			V1: "v2",
		},
		{
			V0: "v2",
		},
	}

	r := New(newTestSource(edges))
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

func TestBinomial_Graph(t *testing.T) {
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
				{V0: "0"},
				{V0: "0", V1: "0"},
				{V0: "0", V1: "0"},
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
			want: []Edge{
				{V0: "0"},
			},
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
			want: []Edge{
				{V0: "0"},
			},
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
				{V0: "0"},
				{V0: "0", V1: "1"},
				{V0: "0", V1: "1"},
				{V0: "1"},
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
				{V0: "0"},
				{V0: "0", V1: "1"},
				{V0: "1"},
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
				{V0: "0"},
				{V0: "0", V1: "1", Directed: true},
				{V0: "1"},
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
			for edge := range b.Graph() {
				got = append(got, edge)
			}

			if !slices.Equal(got, tt.want) {
				t.Errorf("unexpected edges: got: %v want: %v", got, tt.want)
			}
		})
	}
}

func TestBinomial_Graph_edgeless(t *testing.T) {
	const numVertices = 5

	b := &Binomial{
		opts: BinomialOpts{
			Vertices: numVertices,
			N:        0,
			P:        0,
		},
		rand: testRand(),
	}

	var got []Edge
	for edge := range b.Graph() {
		if edge.V1 != "" {
			t.Errorf("unexpected edge: vertex %v is connected with %v", edge.V0, edge.V1)
		}
		got = append(got, edge)
	}

	if len(got) != numVertices {
		t.Errorf("unexpected number of vertices: got: %v want: %v", len(got), numVertices)
	}
}

func TestBinomial_Graph_order_zero(t *testing.T) {
	b := &Binomial{
		opts: BinomialOpts{
			Vertices: 0,
			N:        1,
			P:        1,
		},
		rand: testRand(),
	}

	var got []Edge
	for edge := range b.Graph() {
		got = append(got, edge)
	}

	if len(got) != 0 {
		t.Errorf("expected an order-zero graph: got %v vertices", len(got))
	}
}

func TestLabel(t *testing.T) {
	tests := []struct {
		labels []string
		n      int
		want   string
	}{
		{
			labels: []string{"A", "B"},
			n:      0,
			want:   "A",
		},
		{
			labels: []string{"A", "B"},
			n:      1,
			want:   "B",
		},
		{
			labels: []string{"A", "B"},
			n:      2,
			want:   "A2",
		},
		{
			labels: []string{"A", "B"},
			n:      5,
			want:   "B5",
		},
		{
			labels: nil,
			n:      2,
			want:   "2",
		},
		{
			labels: []string{},
			n:      5,
			want:   "5",
		},
	}
	for _, tt := range tests {
		got := label(tt.labels, tt.n)
		if got != tt.want {
			t.Errorf("unexpected label: got: %q want: %q", got, tt.want)
		}
	}
}

type testSource struct {
	edges []Edge
}

func newTestSource(edges []Edge) *testSource {
	return &testSource{edges: edges}
}

func (src testSource) Graph() <-chan Edge {
	ch := make(chan Edge)
	go func() {
		for _, edge := range src.edges {
			ch <- edge
		}
		close(ch)
	}()
	return ch
}

func testRand() *rand.Rand {
	return rand.New(rand.NewPCG(1, 2))
}
