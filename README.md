pfuzz is a web fuzzer inspired by [ffuf](https://github.com/ffuf/ffuf),
which outputs the generated requests to stdout in the [httpipe
format](https://github.com/codesoap/httpipe) instead of sending them.

# Examples
```console
$ # Fuzzing paths with a wordlist:
$ pfuzz -w /path/to/wordlist -u https://foo.io:1234/FUZZ
{"host":"foo.io","port":"1234","req":"GET /api HTTP/1.1\r\nHost: foo.io:1234\r\n\r\n","tls":true}
{"host":"foo.io","port":"1234","req":"GET /login HTTP/1.1\r\nHost: foo.io:1234\r\n\r\n","tls":true}
{"host":"foo.io","port":"1234","req":"GET /home HTTP/1.1\r\nHost: foo.io:1234\r\n\r\n","tls":true}
...

$ # Using words from stdin to fuzz the Authorization header:
$ generate-tokens | pfuzz -w - -u http://foo.io -H 'Authorization: Bearer FUZZ'
{"host":"foo.io","req":"GET / HTTP/1.1\r\nHost: foo.io\r\nAuthorization: Bearer abc123\r\n\r\n","tls":false}
{"host":"foo.io","req":"GET / HTTP/1.1\r\nHost: foo.io\r\nAuthorization: Bearer xyz1337\r\n\r\n","tls":false}
...

$ # Using multiple wordlists to fuzz paths accross multiple subdomains:
$ pfuzz -w /path/to/subdomains:SUB -w /path/to/paths:PATH -u http://SUB.foo.io/PATH
{"host":"doc.foo.io","req":"GET /api HTTP/1.1\r\nHost: doc.foo.io\r\n\r\n","tls":false}
{"host":"doc.foo.io","req":"GET /login HTTP/1.1\r\nHost: doc.foo.io\r\n\r\n","tls":false}
{"host":"doc.foo.io","req":"GET /home HTTP/1.1\r\nHost: doc.foo.io\r\n\r\n","tls":false}
{"host":"forum.foo.io","req":"GET /api HTTP/1.1\r\nHost: forum.foo.io\r\n\r\n","tls":false}
...
```

# Usage
```console
$ pfuzz -h
Usage of pfuzz:
  -H value
        A HTTP header to use, e.g. 'Content-Type: application/json'.
  -X string
        The HTTP method to use. (default "GET")
  -d string
        Payload data as given, without any encoding.
        Mostly used for POST requests.
  -u string
        The URL of the target.
  -w value
        The path to a wordlist, and optionally a colon followed
        by a custom placeholder, e.g. '/path/to/username/list:USER'.

Zero, one or more wordlists can be provided. If no custom placeholder
is given, FUZZ is used instead; if multiple wordlists have no custom
placeholder, FUZZ2, FUZZ3, etc. will be assigned. If multiple wordlists
are used, all permutations will be generated.

One wordlist can use '-' instead of a path. It's words will be read from
standard input.

If no wordlist is used, only one request will be generated.
```

# TODO
- Maybe allow overwriting the generated `Host` header.
- Maybe allow overwriting the generated `Content-Length` header.
