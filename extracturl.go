package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {

	fin, closefn, ferr, closeferr, err := openSomeFile()
	if err != nil {
		log.Fatal(err)
	}
	defer closefn()
	defer closeferr()

	scanAllines(fin, ferr, combinedLogLineParser)
}

// scanAllines calls a function (argument fn) on all lines
// of fin argument one at a time. Can print some error messages
// on os.Stderr.
func scanAllines(fin *os.File, ferr *os.File, fn func(string) error) {

	scanner := bufio.NewScanner(fin)
	/* For longer lines:
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	*/

	lineCounter := 0

	for scanner.Scan() {
		lineCounter++
		line := scanner.Text()
		if err := fn(line); err != nil {
			if ferr != nil {
				fmt.Fprintf(ferr, "%s\n", line)
			}
			fmt.Fprintf(os.Stderr, "line %d: %v\n", lineCounter, err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "problem line %d: %v", lineCounter, err)
	}
}

// openSomeFile either open a file named by os.Args[1],
// and return an *os.File, or if that command line argument doesn't
// exist, return os.Stdin. Also return a closing function, which should
// always get called, even if openSomeFile returns os.Stdin
func openSomeFile() (*os.File, func(), *os.File, func(), error) {

	badLineFileName := flag.String("b", "", "unparseable lines file name")

	flag.Parse()

	var ferr *os.File
	var closeferr = func() {}
	var err error
	var fn = func() {}

	if *badLineFileName != "" {
		ferr, err = os.OpenFile(*badLineFileName, os.O_CREATE|os.O_WRONLY, 0666)
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

func combinedLogLineParser(textIn string) error {
	// a few user agents contain literal '\"' 2-character strings
	text := strings.ReplaceAll(textIn, `\"`, "''")
	matches := logLineXX.FindAllStringSubmatch(text, -1)
	if len(matches) > 0 {
		// matches[0][1]  IP address
		// matches[0][2]  some garbage
		// matches[0][3]  timestamp
		// matches[0][4]  Method URL HTTPversion
		// matches[0][5]  HTTP status code
		// matches[0][6]  count of bytes sent
		// matches[0][7]  referrer
		// matches[0][8]  User Agent
		fields := strings.Fields(matches[0][4])
		if len(fields) > 1 {
			fmt.Printf("%s\n", fields[1])
		}
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
		return errors.New("no matches")
	}
	return nil
}
