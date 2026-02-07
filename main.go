package main

import (
	"bufio"
	"combined/parser"
	"combined/tree"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

func main() {

	matching, matchingProgram, outputFields, wholeLineOut, rfc3339Timestamps, badLinesFileName, err := examineArguments()
	if err != nil {
		fmt.Fprintf(os.Stderr, "argument error: %v\n", err)
		return
	}

	fopen, err := newFileOpener(badLinesFileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "input file problem: %v\n", err)
		return
	}

	var cont bool

	for cont, err = fopen.NextFile(); cont && err == nil; cont, err = fopen.NextFile() {
		if err := scanAllines(
			fopen.currentFile, fopen.badLinesFile,
			matching, matchingProgram,
			outputFields, wholeLineOut, rfc3339Timestamps); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	err = fopen.Done()
	if err != nil {
		fmt.Fprintf(os.Stderr, "closing files: %v\n", err)
	}

}

// scanAllines calls a function (argument fn) on all lines
// of linesIn argument one at a time. Can print some error messages
// on os.Stderr. Control flow for line scanning.
func scanAllines(linesIn *os.File, linesError *os.File, matching *matchSpec, matchProgram *tree.Node, outputFields []int, wholeLineOut, rfc3339Timestamps bool) error {

	scanner := bufio.NewScanner(linesIn)
	/* For longer lines:
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	*/

	lineCounter := 0

	for scanner.Scan() {
		lineCounter++
		line := scanner.Text()
		if pe, err := combinedLogLineParser(line); err != nil {
			if linesError != nil {
				_, _ = fmt.Fprintf(linesError, "%s\n", line)
			}
			fmt.Fprintf(os.Stderr, "line %d: %v\n", lineCounter, err)
		} else if pe != nil {
			// pe points to a filled-in parsedEntry struct
			if lineMatches(matching, matchProgram, pe) {
				if wholeLineOut {
					fmt.Printf("%s\n", line)
					continue
				}
				performOutput(outputFields, pe, rfc3339Timestamps)
			}
		} else {
			fmt.Fprintf(os.Stderr, "line %d: no error, also no parsed line\n", lineCounter)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("problem line %d: %v", lineCounter, err)
	}

	return nil
}

//	fopen, closefn, closeferr, err :=

type fopenr struct {
	currentFile  *os.File
	badLinesFile *os.File
	nargsIndex   int
}

func (fop *fopenr) Done() error {
	e1 := fop.currentFile.Close()
	if fop.badLinesFile != nil {
		e2 := fop.badLinesFile.Close()
		e1 = errors.Join(e1, e2)
	}
	return e1
}

// NextFile opens another file from list of file names on
// command line after any arguments, if there is one.
// Vacuous case: opens stdin, but only once.
// Returns a bool (true means there's freshly opened file),
// or false and an error. An iterator of sorts.
func (fop *fopenr) NextFile() (bool, error) {
	if flag.NArg() > fop.nargsIndex {
		fop.currentFile.Close()
		var fin *os.File
		var err error
		if fin, err = os.Open(flag.Arg(fop.nargsIndex)); err != nil {
			return false, err
		}
		fop.currentFile = fin
		fop.nargsIndex++
		return true, nil
	} else if fop.currentFile == nil {
		fop.currentFile = os.Stdin
		return true, nil
	}
	return false, nil
}

func newFileOpener(badLinesFileName string) (*fopenr, error) {
	var ferr *os.File
	var err error

	if badLinesFileName != "" {
		ferr, err = os.OpenFile(badLinesFileName, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
	}

	fop := &fopenr{
		currentFile:  nil,
		badLinesFile: ferr,
		nargsIndex:   0,
	}

	return fop, nil
}

var logLineTS = regexp.MustCompile(`^([^ ]+) - ([^ ]*) (\[[^]]+\]).*`)
var logLineUR = regexp.MustCompile(`^([^ ]+) - ([^ ]*) (\[[^]]+\]) "([^"]*)".*`)
var logLineCD = regexp.MustCompile(`^([^ ]+) - ([^ ]*) (\[[^]]+\]) "([^"]*)" (\d{1,}).*`)
var logLineSZ = regexp.MustCompile(`^([^ ]+) - ([^ ]*) (\[[^]]+\]) "([^"]*)" (\d{1,}) (\d{1,}).*`)
var logLineRF = regexp.MustCompile(`^([^ ]+) - ([^ ]*) (\[[^]]+\]) "([^"]*)" (\d{1,}) (\d{1,}) "([^"]*)".*$`)
var logLineXX = regexp.MustCompile(`^([^ ]+) - ([^ ]*) (\[[^]]+\]) "([^"]*)" (\d{1,}) (\d{1,}) "([^"]*)" "([^"]*)"$`)

// parsedEntry holds a combined format line broken into sub-strings. No
// intra-field parsing or interpretation except for method/URL/HTTP version
// field
type parsedEntry struct {
	line   string   // original, entire log file line
	fields []string // different fields in a slice
	// [0]  IP address
	// [1]  some garbage
	// [2]  timestamp
	// [3]  Method
	// [4]  URL
	// [5]  HTTPversion
	// [6]  HTTP status code
	// [7]  count of bytes sent
	// [8]  referrer
	// [9]  User Agent
}

// combinedLogLineParser uses an elaborate regexp to parse
// each line of text it's given into various fields, each of
// which has some semantic content.
func combinedLogLineParser(textIn string) (*parsedEntry, error) {
	// a few user agents contain literal '\"' 2-character strings
	text := strings.ReplaceAll(textIn, `\"`, "''")
	matches := logLineXX.FindAllStringSubmatch(text, -1)
	if len(matches) > 0 {
		if len(matches[0]) > 8 {
			// matches[0][1]  IP address
			// matches[0][2]  some garbage
			// matches[0][3]  timestamp
			// matches[0][4]  Method URL HTTPversion
			// matches[0][5]  HTTP status code
			// matches[0][6]  count of bytes sent
			// matches[0][7]  referrer
			// matches[0][8]  User Agent
			fields := strings.Fields(matches[0][4])
			var method, url, version string
			if len(fields) > 2 {
				method = fields[0]
				url = fields[1]
				version = fields[2]
			}

			p := &parsedEntry{
				line: textIn,
				fields: []string{
					matches[0][1],
					matches[0][2],
					matches[0][3],
					method,
					url,
					version,
					matches[0][5],
					matches[0][6],
					matches[0][7],
					matches[0][8],
				},
			}

			return p, nil
		}
		// line matched the combined format regexp, but somehow
		// did not have the required number of sub-matches. Weird.
		return nil, errors.New("something very bad happened. WTF?")
	} else {
		if matches := logLineTS.FindAllStringSubmatch(text, -1); len(matches) > 0 {
			fmt.Fprintf(os.Stderr, "matched IP address and Timestamp\n")
		}
		if matches := logLineUR.FindAllStringSubmatch(text, -1); len(matches) > 0 {
			fmt.Fprintf(os.Stderr, "matched IP address and Timestamp and URL\n")
		}
		if matches := logLineCD.FindAllStringSubmatch(text, -1); len(matches) > 0 {
			fmt.Fprintf(os.Stderr, "matched IP address and Timestamp and URL and Code\n")
		}
		if matches := logLineSZ.FindAllStringSubmatch(text, -1); len(matches) > 0 {
			fmt.Fprintf(os.Stderr, "matched IP address and Timestamp and URL and Code and Size\n")
		}
		if matches := logLineRF.FindAllStringSubmatch(text, -1); len(matches) > 0 {
			fmt.Fprintf(os.Stderr, "matched IP address and Timestamp and URL and Code and Size and Referer\n")
		}
		return nil, errors.New("no matches")
	}
}

type matchSpec struct {
	matchField  string
	fieldIndex  int
	exactValue  string
	matchRegexp *regexp.Regexp
}

func examineArguments() (*matchSpec, *tree.Node, []int, bool, bool, string, error) {
	badLineFileName := flag.String("b", "", "unparseable lines file name")
	outputFields := flag.String("f", "", "output field(s), comma separated")
	matchExpression := flag.String("m", "", "match expression, field=value or field~regexp")
	wholeLineOutput := flag.Bool("L", false, "output log file line on match, otherwise fields")
	rfc3339Timestamps := flag.Bool("r", false, "output timestamps in RFC3339 format")
	matchProgram := flag.String("e", "", "AND/OR/NOT boolean sentence for match")

	flag.Parse()
	var err error

	var me *matchSpec
	if *matchExpression != "" {
		me, err = createMatching(*matchExpression)
		if err != nil {
			return nil, nil, nil, false, false, "", err
		}
	}

	var ms *tree.Node
	if *matchProgram != "" {
		ms, err = createMatchProgram(*matchProgram)
		if err != nil {
			return nil, nil, nil, false, false, "", err
		}
	}

	return me, ms, createOutputIndexes(*outputFields), *wholeLineOutput, *rfc3339Timestamps, *badLineFileName, nil
}

// createMatching fills in a *matchSpec struct based on
// a "match expression" which is either:
// fieldname=exactstring
// or
// fieldname~regexp
func createMatching(matchExpression string) (*matchSpec, error) {
	if matchExpression == "" {
		return nil, nil
	}
	fields := strings.Split(matchExpression, "=")
	if len(fields) == 2 {
		// exact match desired
		fieldIndex, ok := parser.FieldToIndex[fields[0]]
		if ok {
			pattern := strings.TrimSuffix(strings.TrimPrefix(fields[1], "/"), "/")
			return &matchSpec{
				matchField: fields[0],
				fieldIndex: fieldIndex,
				exactValue: pattern,
			}, nil
		}
		// unknown field
		return nil, fmt.Errorf("unknown input field for exact match %q", fields[0])
	} else {
		fields = strings.Split(matchExpression, "~")
		if len(fields) == 2 {
			// regular expression match desired
			fieldIndex, ok := parser.FieldToIndex[fields[0]]
			if ok {
				pattern := strings.TrimPrefix(strings.TrimSuffix(fields[1], "/"), "/")
				r, err := regexp.Compile(pattern)
				if err != nil {
					return nil, fmt.Errorf("regular expression to match field %q problem: %v", fields[0], fields[1])
				}
				return &matchSpec{
					matchField:  fields[0],
					fieldIndex:  fieldIndex,
					matchRegexp: r,
				}, nil
			}
			return nil, fmt.Errorf("unknown input field for regex match %q", fields[0])
		}
	}
	return nil, fmt.Errorf("bad match spec %q", matchExpression)
}

// lineMatches decides whether a given line of input (broken
// into field as a *parsedEntry) matches the desired criteria.
func lineMatches(ms *matchSpec, mp *tree.Node, pe *parsedEntry) bool {
	if ms == nil && mp == nil {
		return true
	}
	switch {
	case mp != nil:
		return Match(mp, pe)
	case ms != nil:
		if ms.exactValue != "" {
			return ms.exactValue == pe.fields[ms.fieldIndex]
		}
		return ms.matchRegexp.MatchString(pe.fields[ms.fieldIndex])
	}
	fmt.Printf("fall thru false\n")
	return false
}

// performOutput
func performOutput(outputFields []int, pe *parsedEntry, rfc3339Timestamps bool) {
	spacer := ""
	for i := range outputFields {
		if rfc3339Timestamps && outputFields[i] == 2 {
			ts, err := time.Parse(`[02/Jan/2006:15:04:05 +0000]`, pe.fields[outputFields[i]])
			if err != nil {
				fmt.Fprintf(os.Stderr, "time parsing: %v\n", err)
				continue
			}
			fmt.Printf("%s%s", spacer, ts.Format(time.RFC3339))
			spacer = "\t"
			continue
		}
		fmt.Printf("%s%s", spacer, pe.fields[outputFields[i]])
		spacer = "\t"
	}
	fmt.Println()
}

func createOutputIndexes(outputFieldsCSV string) []int {
	if outputFieldsCSV == "" {
		return parser.AllFieldsIndexes
	}

	var indexes []int
	fields := strings.Split(outputFieldsCSV, ",")
	for i := range fields {
		if n, ok := parser.FieldToIndex[fields[i]]; ok {
			indexes = append(indexes, n)
		} else {
			// unknown field
			fmt.Fprintf(os.Stderr, "ignoring unknown output field %q\n", fields[i])
			continue
		}
	}
	sort.Ints(indexes)
	return indexes
}
