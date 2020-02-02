package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"strconv"
)

const pgmname = "match"

type optT struct {
	nodup bool
	outs  []string
}

type inT struct {
	name   string // filename
	count  int    // read lines
	rec    string // current record
	eof    bool   // current read status
	eod    bool   // current get status
	file   *os.File
	reader *bufio.Reader
}

type outT struct {
	name   string // filename
	count  int    // written lines
	file   *os.File
	writer *bufio.Writer
}

type allT struct {
	pats  []string // all possible patterns
	nodup bool     // ignore duplicate lines
	ins   []inT    // input files
	outs  []outT   // output files
}

// Flip is a helper.
func flip(ss []string) {
	for i, s := range ss {
		if i+1 > len(ss)/2 {
			break
		}
		ss[i] = ss[len(ss)-1-i]
		ss[len(ss)-1-i] = s
	}
}

// Command match handles arbitrary input files, compares them line by line,
// and creates new files whose names reflect the match/nomatch patterns
// and contain matched lines respectively. Input files must be sorted.
func main() {
	repFile := mustCreate("Report.txt")
	defer repFile.Close()
	repWriter := bufio.NewWriter(repFile)
	defer repWriter.Flush()
	reportHead(repWriter)

	var all allT

	filenames, opts := args()
	flip(filenames) // so that position in arguments matches position in matching pattern

	all.pats = allPats(len(filenames)) // init all possible patterns

	if opts.nodup { // no duplicates
		all.nodup = opts.nodup
	}

	all.ins = make([]inT, 0, len(filenames)) // input files
	for _, fn := range filenames {
		var in inT
		in.name = fn
		in.file = mustOpen(in.name)
		defer in.file.Close()
		all.ins = append(all.ins, in)
	}

	all.outs = make([]outT, 0, len(all.pats)-1) // output files (minus Ns only pattern)
	for i, pat := range all.pats {
		if i == 0 { // no file for Ns only pattern
			continue
		}
		var out outT
		out.name = pat
		if len(opts.outs) > 0 {
			for _, name := range opts.outs {
				if pat == name { // desired output files only
					out.file = mustCreate(out.name)
					defer out.file.Close()
				}
			}
		} else { // all possible output files
			out.file = mustCreate(out.name)
			defer out.file.Close()
		}
		all.outs = append(all.outs, out)
	}

	err := process(&all) // ACTION
	if err != nil {
		log.Fatalln(err)
	}

	reportFoot(repWriter, &all)

	return
}

func reportHead(report *bufio.Writer) {
	var cmd string
	for _, v := range os.Args {
		cmd = cmd + v + " "
	}
	mustWrite(report, "Command:\n\t$ "+cmd)
}

func reportFoot(report *bufio.Writer, all *allT) {
	mustWrite(report, "\nInput files:")
	for i := len(all.ins) - 1; i >= 0; i-- {
		v := all.ins[i]
		mustWrite(report, "\t"+v.name+": "+strconv.Itoa(v.count))
	}
	mustWrite(report, "\nOutput files:")
	for _, v := range all.outs {
		mustWrite(report, "\t"+v.name+": "+strconv.Itoa(v.count))
	}
}

// AllPats returns all possible Y/N patterns with length len.
// Indices of slice pats correspond to the contained patterns,
// i.e. you get the Index if you replace all N with 0 and all Y with 1
// and read the result as as binary number.
//
// Example: for len 2 allPats return pats = {"NN" "NY" "YN" "YY"}.
func allPats(len int) (pats []string) {
	if len == 0 { // no pattern
		return nil
	}

	pats = make([]string, 0, (1 << len))

	for i := 0; i < (1 << len); i++ { // 1<<len == 2**len
		pat, val := "", i
		for pos := len; pos > 0; pos-- {
			if val >= (1 << (pos - 1)) { // 1<<(pos-1) == 2**(pos-1)
				pat += "Y"
				val -= (1 << (pos - 1))
			} else {
				pat += "N"
			}
		}
		pats = append(pats, pat)
	}
	return pats
}

// Process creates bufio readers and writers and calls matchAll for action.
func process(all *allT) (err error) {
	for i, in := range all.ins {
		all.ins[i].reader = bufio.NewReader(in.file)
	}

	for i, out := range all.outs {
		if out.file != nil { // desired output files only
			all.outs[i].writer = bufio.NewWriter(out.file)
			defer all.outs[i].writer.Flush()
		}
	}

	err = matchAll(all) // ACTION
	if err != nil {
		log.Fatalln(err)
	}

	return err
}

// MatchAll reads all input in a read loop and writes to the output files.
func matchAll(all *allT) (err error) {

	for i, _ := range all.ins { // initial readings
		err = get(i, all)
		if err != nil {
			return err
		}
	}
	curPatInx, inInxs := compare(all)

	for curPatInx != 0 {
		for patInx := range all.pats {
			if curPatInx == patInx { // MATCH

				if all.outs[patInx-1].writer != nil { // desired output files only
					// [patInx-1], because pattern with Ns only makes no output file
					mustWrite(all.outs[patInx-1].writer, all.ins[inInxs[0]].rec)
					all.outs[patInx-1].count += 1
				}

				for _, i := range inInxs { // read input files where lines were processed
					err = get(i, all)
					if err != nil {
						return err
					}
				}
				curPatInx, inInxs = compare(all)
				break
			}
		}
	}

	return nil
}

// Compare compares all input lines currently in focus,
// returns the pattern index for the minimum line and all indices
// that lead to the input files whose current lines are to be processed.
func compare(all *allT) (curPatInx int, inInxs []int) {
	if allEOD(all) {
		return 0, nil
	}

	var recs = make([]string, 0, len(all.ins))
	for i := range all.ins {
		recs = append(recs, all.ins[i].rec)
	}
	minRec := min(recs...)

	for i := range all.ins {
		if all.ins[i].rec == minRec {
			curPatInx |= (1 << i)      // für Mustervergleich
			inInxs = append(inInxs, i) // am Muster Beteiligte
		}
	}

	return curPatInx, inInxs
}

func allEOD(all *allT) bool {
	for i := range all.ins {
		if !all.ins[i].eod {
			return false
		}
	}
	return true
}

func min(ss ...string) (s string) {
	s = ss[0]
	for _, val := range ss {
		if val < s {
			s = val
		}
	}
	return s
}

// Get reads the next line from all.ins[i].reader and feeds the
// variables all.ins[i].eof/eod/rec.
// With option -nodup it ignores duplicate lines, and empty ones too.
// If an error occurs get stops working and returns an error value.
func get(i int, all *allT) error {
	if all.ins[i].eof { // earlier: EOF + data
		all.ins[i].eod = true
		all.ins[i].rec = string(byte(0xFF))
		return nil
	}

readagain:
	rec, err := all.ins[i].reader.ReadString('\n')

	if err != nil && err != io.EOF { // serious read error
		log.Fatalln(all.ins[i].name+": read error after line:", all.ins[i].rec, "\n\t", err)
	}

	if err == io.EOF && len(rec) == 0 { // EOF + no data
		all.ins[i].eof = true
		all.ins[i].eod = true
		all.ins[i].rec = string(byte(0xFF))
		return nil
	}

	if err == io.EOF && len(rec) > 0 { // EOF + data
		all.ins[i].eof = true
	}

	if rec[len(rec)-1] == '\n' {
		rec = rec[:len(rec)-1] // data without delimiter
	}

	all.ins[i].count += 1

	// -nodup ignores duplicate lines (and empty ones)
	if all.nodup && rec == all.ins[i].rec {
		goto readagain
	}

	// lines must be ordered ascendingly (compare with previous line)
	if rec < all.ins[i].rec {
		return errors.New(all.ins[i].name + ": wrong sequence near:\n" + all.ins[i].rec)
	}

	all.ins[i].rec = rec

	return nil
}

func mustOpen(name string) (f *os.File) {
	f, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func mustCreate(name string) (f *os.File) {
	f, err := os.Create(name)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// MustWrite writes rec and ends the line.
func mustWrite(f *bufio.Writer, rec string) {
	_, err := f.WriteString(rec)
	if err != nil {
		log.Fatal(err)
	}
	err = f.WriteByte('\n')
	if err != nil {
		log.Fatal(err)
	}
}
