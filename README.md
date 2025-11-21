# Apache httpd combined log format field extractor

Ever wanted to peel out just URL or user agent or timestamps
from [Apache httpd]()
"combined" format log files?
I have, so I wrote this program

## Examples

```
$ ./combined -f ipaddr,timestamp /var/log/httpd/access_log
```

## Field names

```
ipaddr
garbage
timestamp
method
url
version
code
size
referrer
useragent
```

