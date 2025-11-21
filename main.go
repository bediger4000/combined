package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

func main() {

	matching, outputFields, badLinesFileName, err := examineArguments()

	fin, closefn, ferr, closeferr, err := openSomeFile(badLinesFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer closefn()
	defer closeferr()

	scanAllines(fin, ferr, matching, outputFields)
}

// scanAllines calls a function (argument fn) on all lines
// of linesIn argument one at a time. Can print some error messages
// on os.Stderr.
func scanAllines(linesIn *os.File, linesError *os.File, matching *matchSpec, outputFields []int) error {

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
				fmt.Fprintf(linesError, "%s\n", line)
			}
			fmt.Fprintf(os.Stderr, "line %d: %v\n", lineCounter, err)
		} else if pe != nil {
			// pe points to a filled-in parsedEntry struct
			if lineMatches(matching, pe) {
				performOutput(outputFields, pe)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("problem line %d: %v", lineCounter, err)
	}

	return nil
}

// openSomeFile either open a file named by flag.Arg(0),
// and return an *os.File, or if that command line argument doesn't
// exist, return os.Stdin. Also return a closing function, which should
// always get called, even if openSomeFile returns os.Stdin
func openSomeFile(badLineFileName string) (*os.File, func(), *os.File, func(), error) {

	var ferr *os.File
	var closeferr = func() {}
	var err error
	var fn = func() {}

	if badLineFileName != "" {
		ferr, err = os.OpenFile(badLineFileName, os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return nil, fn, nil, closeferr, err
		}
		closeferr = func() { ferr.Close() }
	}

	fin := os.Stdin
	if flag.NArg() > 0 {
		var err error
		if fin, err = os.Open(flag.Arg(0)); err != nil {
			return nil, fn, nil, closeferr, err
		}
		fn = func() { fin.Close() }
	}
	return fin, fn, ferr, closeferr, nil
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
	return nil, nil
}

type matchSpec struct {
	matchField  string
	fieldIndex  int
	exactValue  string
	matchRegexp *regexp.Regexp
}

func examineArguments() (*matchSpec, []int, string, error) {
	badLineFileName := flag.String("b", "", "unparseable lines file name")
	outputFields := flag.String("f", "", "output field(s), comma separated")
	matchExpression := flag.String("m", "", "match expression, field=value or field~regexp")

	flag.Parse()

	me, err := createMatching(*matchExpression)
	if err != nil {
	}

	return me, createOutputIndexes(*outputFields), *badLineFileName, nil
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
		fieldIndex, ok := fieldToIndex[fields[0]]
		if ok {
			return &matchSpec{
				matchField: fields[0],
				fieldIndex: fieldIndex,
				exactValue: fields[1],
			}, nil
		}
		// unknown field
		return nil, fmt.Errorf("unknown input field for exact match %q", fields[0])
	} else {
		fields = strings.Split(matchExpression, "~")
		if len(fields) == 2 {
			// regular expression match desired
			fieldIndex, ok := fieldToIndex[fields[0]]
			if ok {
				r, err := regexp.Compile(fields[1])
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
func lineMatches(ms *matchSpec, pe *parsedEntry) bool {
	switch {
	case ms == nil:
		return true
	case ms.exactValue != "":
		if ms.exactValue == pe.fields[ms.fieldIndex] {
			return true
		}
	case ms.matchRegexp != nil:
		return ms.matchRegexp.MatchString(pe.fields[ms.fieldIndex])
	}
	return false
}

// performOutput
func performOutput(outputFields []int, pe *parsedEntry) {
	spacer := ""
	for i := range outputFields {
		fmt.Printf("%s%s", spacer, pe.fields[outputFields[i]])
		spacer = "\t"
	}
	fmt.Println()
}

var fieldToIndex = map[string]int{
	"ipaddr":    0,
	"garbage":   1,
	"timestamp": 2,
	"method":    3,
	"url":       4,
	"version":   5,
	"code":      6,
	"size":      7,
	"referrer":  8,
	"useragent": 9,
}

func createOutputIndexes(outputFieldsCSV string) []int {
	if outputFieldsCSV == "" {
		return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	}

	var indexes []int
	fields := strings.Split(outputFieldsCSV, ",")
	for i := range fields {
		if n, ok := fieldToIndex[fields[i]]; ok {
			indexes = append(indexes, n)
		} else {
			// unknown field
			fmt.Fprintf(os.Stderr, "unknown output field %q\n", fields[i])
			continue
		}
	}
	sort.Ints(indexes)
	return indexes
}
