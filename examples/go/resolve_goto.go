// Resolve a Google /goto URL by reading the Location header.
//
// Google Search results often wrap destinations as:
//
//	https://www.google.com/goto?url=CAES...
//
// The CAES blob is opaque — you cannot decode it. Google only reveals the
// real URL in the HTTP Location header.
//
// Do not follow the redirect (CheckRedirect returns ErrUseLastResponse).
// Try HEAD first, then GET. Google sometimes returns 402 (or another
// non-3xx) with Location: still read the header.
//
// Usage:
//
//	go run ./examples/go 'https://www.google.com/goto?url=CAES...'
package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func isGoogleGoto(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return u.Scheme == "https" &&
		(host == "google.com" || host == "www.google.com") &&
		(u.Path == "/goto" || strings.HasPrefix(u.Path, "/goto/"))
}

func resolveGoto(raw string) (string, error) {
	if !isGoogleGoto(raw) {
		return "", errors.New("expected https://www.google.com/goto?url=...")
	}

	// Default Client follows 3xx; stop on the first hop and read Location.
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for _, method := range []string{http.MethodHead, http.MethodGet} {
		req, err := http.NewRequest(method, raw, nil)
		if err != nil {
			return "", err
		}
		// Referer matches a real search click; some hops expect it.
		req.Header.Set("Referer", "https://www.google.com/search")

		res, err := client.Do(req)
		if err != nil {
			return "", err
		}
		location := res.Header.Get("Location")
		_ = res.Body.Close()
		if location == "" {
			continue
		}
		// Location may be relative; resolve it against the request URL.
		abs, err := res.Request.URL.Parse(location)
		if err != nil {
			return "", err
		}
		return abs.String(), nil
	}
	return "", errors.New("no Location header")
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "Usage: go run . 'https://www.google.com/goto?url=CAES...'")
		os.Exit(2)
	}
	dest, err := resolveGoto(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(dest)
}
