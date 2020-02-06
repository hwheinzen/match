// Datei match_test.go enthält Tests.

package main

import (
	"bytes"
	"strconv"
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

func TestCompare(t *testing.T) {
	fncname := "TestCompare"
	tests := []struct {
		in   *allT
		patInx int
		inInxs []int
	}{
		{nil, 0, nil},
		{&allT{}, 0, nil},
		{&allT{ins: []inT{inT{eod: true}}}, 0, nil},
		{&allT{ins: []inT{inT{rec: "A"}}}, 1, []int{0}}, // pattern 1: Y
		{&allT{ins: []inT{inT{rec: "A"}, inT{rec: "B"}}}, 1, []int{0}}, // pattern 1: NY
		{&allT{ins: []inT{inT{rec: "B"}, inT{rec: "A"}}}, 2, []int{1}}, // pattern 2: YN
		{&allT{ins: []inT{inT{rec: "A"}, inT{rec: "A"}}}, 3, []int{0, 1}}, // pattern 3: YY
		{&allT{ins: []inT{inT{rec: "A"}, inT{rec: "A"}, inT{rec: "B"}}}, 3, []int{0, 1}}, // pattern 3: NYY
		{&allT{ins: []inT{inT{rec: "B"}, inT{rec: "A"}, inT{rec: "A"}}}, 6, []int{1, 2}}, // pattern 6: YYN
	}
	for _, v := range tests {
		pi, is := compare(v.in)
		if pi != v.patInx {
			t.Error(fncname, "wanted:", v.patInx, v.inInxs, "got:", pi, is)
		} else if is == nil {
			if v.inInxs != nil {
				t.Error(fncname, "wanted:", v.patInx, v.inInxs, "got:", pi, is)
			} else {
				for i, inx := range is {
					if inx != v.inInxs[i] {
						t.Error(fncname, "wanted:", v.patInx, v.inInxs, "got:", pi, is)
					}
				}
			}
		}
	}
}

// Vorbereiten entsprechend main:
// - var all
// - für all.ins: jeweils bytes.NewBufferString
// - allPats(len(all.ins))
// - für all.outs: jeweils bytes.NewBufferString
// - range tests
// - - process
func TestProcess(t *testing.T) {
	fncname := "TestProcess"
	tests := []struct {
		ins  []string
		outs []string
	}{
		{
			[]string{ // 1 Eingabe
`a
aa
aaa
`,
			},
			[]string{ // 1 Ausgabe erwartet
`a
aa
aaa
`, // Y
			},
		},
		{
			[]string{ // 3 Eingaben
`a
ab
abc
ac
`, // erste
`ab
abc
b
bc
`, // zweite
`abc
ac
bc
c
cc
`, // dritte
			},
			[]string{ // 7 Ausgaben erwartet
`a
`, // NNY
`b
`, // NYN
`ab
`, // NYY
`c
cc
`, // YNN
`ac
`, // YNY
`bc
`, // YYN
`abc
`, // YYY
			},
		},
	}
	for _, v := range tests {
		var all allT

		all.ins = make([]inT, 0, len(v.ins))
		for i := 0; i < len(v.ins); i++ {
			var in inT
			in.name = strconv.Itoa(i)
			buf := bytes.NewBufferString(v.ins[i])
			in.ioreader = buf
			all.ins = append(all.ins, in)
		}

		all.pats = allPats(len(v.ins))

		all.outs = make([]outT, 0, len(all.pats)-1)
		for i, pat := range all.pats {
			if i == 0 { // no file for Ns only pattern
				continue
			}
			var out outT
			out.name = pat
			buf := bytes.NewBuffer([]byte{})
			out.iowriter = buf
			all.outs = append(all.outs, out)
		}

		err := process(&all)
		if err != nil {
			t.Fatal(fncname+":", err)
		}

		for i, out := range all.outs {
			if out.iowriter.(*bytes.Buffer).String() != v.outs[i] {
				t.Error(fncname, "wanted:", v.outs[i], "got:", out.iowriter.(*bytes.Buffer).String())
			}
		}
		
	}
}
