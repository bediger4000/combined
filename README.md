# Apache httpd combined log format field extractor

Ever wanted to peel out just URL or user agent or timestamps
from [Apache httpd](https://httpd.apache.org/)
"combined" format log files?
I have, so I wrote this program.

I also wanted to combine criteria logically,
say "give me IP addresses of any log file entries that retrieve a `/posts/` URL,
and have a `slashdot.org` referral.

I wrote a complicated Go [regexp](https://github.com/google/re2/wiki/Syntax)
that recognizes an Apache "combined" log file format line,
breaking the line into fields.

I wanted to be able to match particular fields via regular expression.

Then, I had the idea of accepting a small propositional logic sentence on the command line,
and evaluating it for every log line.
The truth values are either exact string matches or regular expression matches
of particular fields.

This program is a mash-up of those two ideas.

## Build

```
$ cd $GOPATH/src
$ git clone git@github.com:bediger4000/combined.git
$ cd combined
$ go test -v ./...
$ go build $PWD
```

## Run

There' two ways of running, one a little faster than the other.

Faster way: exact match or regular expression match of one field:

```
$ combined -f ipaddr,timestamp -r -m 'url=/exact string to match/' /var/log/httpd/access_log
$ combined -f url,referrer -m 'url~/..*\.php/' /var/log/httpd/access_log /var/log/httpd/access_log.1
```

Slower, but more expressive, way.
A propositional logic sentence that gets evaluated on each input line.

```
$  cat /var/log/httpd/access.log.* |
    combined  -e 'ipaddr~/192\.243\..*/ && url~/..*html?..*/ && method=/POST/` -r -f timestamp,url
```

That will print timestamp in RFC3339 format and URL of all POST requests from IPv4 addresses
beginning with '192.243.', and asking for an HTML file with name/value parameters on the end.
It prints the URL asked for, and the timestamp.

### Command Line Flags

```
  -L    output log file line on match, otherwise fields
  -b string
        unparseable lines file name
  -e string
        AND/OR/NOT boolean sentence for match
  -f string
        output field(s), comma separated
  -m string
        match expression, field=value or field~regexp
  -r    output timestamps in RFC3339 format
```

The `-m` flag prints lines that have a field that
matches a string exactly, or a field matches a regular expression.

### Matching expressions

Both single field matches, and logical sentence matchs have the same syntax:

- &lt;field&gt; = /matching string/
- &lt;field&gt; ~ /regular expression/

### Logical sentences

Logical sentence matching has the same matching specifications,
but the matching specifications can be connected with `&&` for "logical AND",
`||` for "logical OR", `-` for NOT, and parenthesized for clarity.
NOT binds tightest, followed by AND, followed by OR.

Here's the grammar for reference:

expr     &rarr; term { OR term }<br/>
term     &rarr; factor { AND factor }<br/>
factor   &rarr; '(' expr ')' | NOT factor | boolean<br/>
boolean  &rarr; FIELD match-op PATTERN<br/>
match-op &rarr; '='|'~'<br/>


#### Field names

The field names get used both in match expressions,
and with the `-f` flag to specify which field to print on the occasion of a match.

|         |
|:--------|
|ipaddr
|garbage
|timestamp
|method
|url
|version
|code
|size
|referrer
|useragent


The `url` field could arguably be called `path`, and I didn't misspell `referrer`.

## Examples

Print iP address and timestamp of every request in a "combined" format log file:
```
$ ./combined -f ipaddr,timestamp /var/log/httpd/access_log
```

Print the URL and referrer for a [Hugo](https://gohugo.io/)
blog that an Apache server provides,
if the referrer is one of the "tags" pages Hugo provides.
```
$ combined -e 'url~/\/posts\/..*/ && referrer~/bruceediger.com\/tags\//' \
    -f url,referrer  slashdot.sorted.human.log 
```

## Design

The program does this:

1. Evaluate command line arguments.
  - Includes parsing any logical expressions set with `-e` flag.
2. Read in a line of text.
3. See if it matches the regular expression for a "combined" format line.
   - Break the line into the logical fields of a "combined" format record
4. If there's a simple exact match or regex match set,
see if the fields match.
5. If there's a logical expression set,
evaluate the expression with the fields.
6. If step (4) or (5) indicate, output the fields specified.
7. Repeat steps 2-6 until no lines are left.

I used the `flag` Go standard package to read the command line.
This lets me use a file name on the command line for input,
or if that's lacking, read log file lines from stdin.

The program reads log file lines with a `bufio.Scanner` struct
from the `bufio` Go standard package.

Matching the log file lines (step 3) gets done with a
moderately complicated `regexp` from the Go standard packages.
 
I kept the logical fields of the "combined" format log file
in a slice of Go's type `string`,
so that the program field could use numerical indexes to find
fields for either matching, or output.

I wrote a recursive descent parser for the logical expressions,
and a single function recursive evalution,
which gets run for each input line if the `-e` command line option
appears.
The evaluation is done on the abstract syntax (parser output) tree,
it's a "big step" evaluation, not compiled to some kind of byte code.
The parsing routines return a `*tree.Node`,
representing the abstract syntax tree, and an error.
Because users specify regular expressions to match fields
of log lines, and entire expressions,
there are more errors from parsing than from evaluation.
That's different from similar evalution of arithmetic expressions.
Once the logical expressions parse correctly,
evaluation doesn't have any errors.
Arithmetical expressions can have evaluation errors
like divide-by-zero, after the expression parses correctly.

