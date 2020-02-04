// Datei match_test.go enthält Tests.

package main

import (
	"testing"
)

func TestFlip(t *testing.T) {
	fncname := "TestFlip"
	tests := []struct {
		inout []string
		want  []string
	}{
		{nil, nil},
		{[]string{""}, []string{""}},
		{[]string{"", ""}, []string{"", ""}},
		{[]string{"", "", ""}, []string{"", "", ""}},
		{[]string{"A"}, []string{"A"}},
		{[]string{"A", "B"}, []string{"B", "A"}},
		{[]string{"A", "B", "C"}, []string{"C", "B", "A"}},
	}
	for _, v := range tests {
		if v.inout == nil {
			flip(v.inout)
			if v.inout != nil {
				t.Error(fncname, "wanted: nil", "got:", v.inout)
				continue
			}
		}
		flip(v.inout)
		for i, s := range v.inout {
			if s != v.want[i] {
				t.Error(fncname, "wanted:", v.want, "got:", v.inout)
			}
		}
	}
}

func TestAllPats(t *testing.T) {
	fncname := "TestAllPats"
	tests := []struct {
		in   int
		want []string
	}{
		{0, nil},
		{1, []string{"N", "Y"}},
		{2, []string{"NN", "NY", "YN", "YY"}},
		{3, []string{"NNN", "NNY", "NYN", "NYY", "YNN", "YNY", "YYN", "YYY"}},
	}
	for _, v := range tests {
		out := allPats(v.in)
		if v.in == 0 {
			if out != nil {
				t.Error(fncname, "wanted:", v.want, "got:", out)
			}
		} else {
			for i, o := range out {
				if o != v.want[i] {
					t.Error(fncname, "wanted:", v.want, "got:", out)
					break
				}
			}
		}
	}
}

func TestMin(t *testing.T) {
	fncname := "TestMin"
	tests := []struct {
		in   []string
		want string
	}{
		{nil, ""},
		{[]string{""}, ""},
		{[]string{"A", ""}, ""},
		{[]string{"B", "A"}, "A"},
		{[]string{"A", "B"}, "A"},
	}
	for _, v := range tests {
		out := min(v.in...)
		if out != v.want {
			t.Error(fncname, "wanted:", v.want, "got:", out)
		}
	}
}

func TestAllEOD(t *testing.T) {
	fncname := "TestAllEOD"
	tests := []struct {
		in   *allT
		want bool
	}{
		{nil, true},
		{&allT{}, true},
		{&allT{ins: []inT{inT{eod: false}}}, false},
		{&allT{ins: []inT{inT{eod: true}}}, true},
		{&allT{ins: []inT{inT{eod: true}, inT{eod: true}}}, true},
		{&allT{ins: []inT{inT{eod: true}, inT{eod: false}}}, false},
		{&allT{ins: []inT{inT{eod: false}, inT{eod: true}}}, false},
	}
	for _, v := range tests {
		out := allEOD(v.in)
		if out != v.want {
			t.Error(fncname, "wanted:", v.want, "got:", out)
		}
	}
}
