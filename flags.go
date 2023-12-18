package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var usageDetails = `
Zero, one or more wordlists can be provided. If no custom placeholder
is given, FUZZ is used instead; if multiple wordlists have no custom
placeholder, FUZZ2, FUZZ3, etc. will be assigned. If multiple wordlists
are used, all permutations will be generated.

One wordlist can use '-' instead of a path. It's words will be read from
standard input.

If no wordlist is used, only one request will be generated.
`

type multiStringFlag []string

func (i multiStringFlag) String() string      { return strings.Join(i, ", ") }
func (i *multiStringFlag) Set(s string) error { *i = append(*i, s); return nil }

func parseFlags() ([]wordlist, string, multiStringFlag, string, string) {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprint(flag.CommandLine.Output(), usageDetails)
	}

	var rawWordlists multiStringFlag
	var rawURL string
	var headers multiStringFlag
	var data string
	var method string
	flag.Var(&rawWordlists, "w", "The path to a wordlist, and optionally a "+
		"colon followed\nby a custom placeholder, e.g. '/path/to/username/list:USER'.")
	flag.StringVar(&rawURL, "u", "", "The URL of the target.")
	flag.Var(&headers, "H", "A HTTP header to use, e.g. 'Content-Type: application/json'.")
	flag.StringVar(&data, "d", "", "Payload data as given, without any encoding. "+
		"\nMostly used for POST requests.")
	flag.StringVar(&method, "X", "GET", "The HTTP method to use.")
	flag.Parse()

	wordlists := parseWordlists(rawWordlists)
	checkHeaders(headers)
	return wordlists, rawURL, headers, data, method
}

func checkHeaders(headers []string) {
	for _, header := range headers {
		if len(header) == 0 {
			fmt.Fprintln(os.Stderr, "Error: Empty headers are not allowed.")
			os.Exit(1)
		}
	}
}

// parseWordlists takes the given wordlist parameters, separates paths
// from placeholders and generates placeholders if needed. The result is
// returned.
//
// If the '-' path is used (for stdin), it will be the first wordlist in
// the returned value.
func parseWordlists(rawWordlists []string) []wordlist {
	var wordlists []wordlist
	usedPlaceholders := make(map[string]bool)
	stdinUsed := false
	generatedPlaceholderNumber := 1
	for _, rawWordlist := range rawWordlists {
		split := strings.Split(rawWordlist, ":")
		if len(split) > 2 {
			fmt.Fprintf(os.Stderr, "Error: Wordlist '%s' contains multiple colons.\n", rawWordlist)
			os.Exit(1)
		}
		var placeholder string
		if len(split) == 2 {
			placeholder = split[1]
			if len(placeholder) == 0 {
				fmt.Fprintf(os.Stderr, "Error: Wordlist '%s' has an empty placeholder.\n", rawWordlist)
				os.Exit(1)
			}
		} else {
			if generatedPlaceholderNumber == 1 {
				placeholder = "FUZZ"
			} else {
				placeholder = fmt.Sprintf("FUZZ%d", generatedPlaceholderNumber)
			}
			generatedPlaceholderNumber++
		}
		if _, alreadyUsed := usedPlaceholders[placeholder]; alreadyUsed {
			fmt.Fprintf(os.Stderr, "Error: The placeholder '%s' cannot be used twice.\n", placeholder)
			os.Exit(1)
		}
		usedPlaceholders[placeholder] = true
		if len(split[0]) == 0 {
			fmt.Fprintf(os.Stderr, "Error: Wordlist '%s' has empty path.\n", rawWordlist)
			os.Exit(1)
		}
		if split[0] == "-" {
			if stdinUsed {
				fmt.Fprintln(os.Stderr, "Error: Standard input can only be used once.")
				os.Exit(1)
			}
			stdinUsed = true
		}
		wordlists = append(wordlists, wordlist{path: split[0], placeholder: placeholder})
	}
	return moveStdinToFront(wordlists)
}

func moveStdinToFront(wordlists []wordlist) []wordlist {
	for i, wl := range wordlists {
		if wl.path == "-" {
			return append(append([]wordlist{wl}, wordlists[:i]...), wordlists[i+1:]...)
		}
	}
	return wordlists
}
