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
$ combined -f url,referrer -m 'url~/..*\.php/' /var/log/httpd/access_log
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

#### Field names


|---------|
| ipaddr |
| garbage |
| timestamp |
| method |
| url |
| version |
| code |
| size |
| referrer |
| useragent |


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

