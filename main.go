package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"slices"
	"strings"
)

type wordlist struct {
	path        string // - if stdin shall be used
	placeholder string
}

func main() {
	wordlists, u, headers, data, method := parseFlags()
	placeholders := extractPlaceholders(wordlists)
	usedPlaceholders := usedPlaceholders(placeholders, u, headers, data)
	usedWordlists := usedWordlists(wordlists, usedPlaceholders)

	placeholderValues := make(chan map[string]string)
	go permutate(usedWordlists, placeholderValues)
	for pvs := range placeholderValues {
		fmt.Println(toOutLine(u, headers, data, method, usedPlaceholders, pvs))
	}
}

func extractPlaceholders(ws []wordlist) []string {
	var placeholders []string
	for _, w := range ws {
		placeholders = append(placeholders, w.placeholder)
	}

	// Sort placeholders from longest to shortest. This will be used
	// for replacement order later; it ensures, that no parts of longer
	// placeholders are replaced by shorter ones.
	slices.SortFunc(placeholders, func(a, b string) int { return len(b) - len(a) })

	return placeholders
}

func usedPlaceholders(placeholders []string, u *url.URL, headers []string, data string) []string {
	// tmp is a combination of all parts that can contain placeholders.
	tmp := u.Hostname() + "\n" +
		u.RequestURI() + "\n" +
		data + "\n"
	for _, header := range headers {
		tmp += header + "\n"
	}

	var usedPlaceholders []string
	prevLen := len(tmp)
	for _, ph := range placeholders {
		tmp = strings.ReplaceAll(tmp, ph, "")
		if len(tmp) < prevLen {
			usedPlaceholders = append(usedPlaceholders, ph)
		}
		prevLen = len(tmp)
	}
	return usedPlaceholders
}

func usedWordlists(wordlists []wordlist, usedPlaceholders []string) []wordlist {
	var usedWordlists []wordlist
	for _, wl := range wordlists {
		for _, usedPlaceholder := range usedPlaceholders {
			if usedPlaceholder == wl.placeholder {
				usedWordlists = append(usedWordlists, wl)
				break
			}
		}
	}
	return usedWordlists
}

func permutate(wordlists []wordlist, permutations chan map[string]string) {
	defer close(permutations)
	if len(wordlists) == 0 {
		permutations <- make(map[string]string)
		return
	}

	var in io.Reader
	filename := wordlists[0].path
	if filename == "-" {
		in = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error: Wordlist could not be opened:", err)
			os.Exit(1)
		}
		defer file.Close()
		in = file
	}

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		word := scanner.Text()
		subPermutations := make(chan map[string]string)
		go permutate(wordlists[1:], subPermutations)
		for subPermutation := range subPermutations {
			subPermutation[wordlists[0].placeholder] = word
			permutations <- subPermutation
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: Reading word from '%s' failed: %v\n", filename, err)
		os.Exit(1)
	}
}

func toOutLine(u *url.URL, headers []string, data, method string, ps []string, pvs map[string]string) string {
	j := make(map[string]any)
	j["host"] = doReplacements(u.Hostname(), ps, pvs)
	if u.Port() != "" {
		j["port"] = u.Port()
	}
	j["tls"] = u.Scheme == "https"
	j["req"] = toRequest(u, headers, data, method, ps, pvs)
	b, err := json.Marshal(j)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error: Could not generate JSON:", err)
		os.Exit(1)
	}
	return string(b)
}

func toRequest(u *url.URL, headers []string, data, method string, ps []string, pvs map[string]string) string {
	var req strings.Builder
	fmt.Fprintf(&req, "%s %s HTTP/1.1\r\n", method, doReplacements(u.RequestURI(), ps, pvs))
	if u.Port() == "" {
		fmt.Fprintf(&req, "Host: %s\r\n", doReplacements(u.Hostname(), ps, pvs))
	} else {
		fmt.Fprintf(&req, "Host: %s:%s\r\n", doReplacements(u.Hostname(), ps, pvs), u.Port())
	}
	for _, header := range headers {
		fmt.Fprintf(&req, "%s\r\n", doReplacements(header, ps, pvs))
	}
	if data == "" {
		fmt.Fprintf(&req, "\r\n")
	} else {
		data = doReplacements(data, ps, pvs)
		fmt.Fprintf(&req, "Content-Length: %d\r\n\r\n%s", len(data), data)
	}
	return req.String()
}

func doReplacements(s string, ps []string, pvs map[string]string) string {
	for _, placeholder := range ps {
		s = strings.ReplaceAll(s, placeholder, pvs[placeholder])
	}
	return s
}
