// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Reads govulncheck's JSON on stdin and fails on anything called that is not
// accepted in allow.txt.
//
// govulncheck reports a vulnerability at module, package and symbol level. Only
// a finding whose trace begins at a function is one this code actually calls;
// the rest are present in the dependency graph and never reached.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type entry struct {
	OSV *struct {
		ID      string `json:"id"`
		Summary string `json:"summary"`
	} `json:"osv"`
	Finding *struct {
		OSV   string `json:"osv"`
		Trace []struct {
			Module   string `json:"module"`
			Function string `json:"function"`
		} `json:"trace"`
	} `json:"finding"`
}

func allowed() (map[string]bool, error) {
	_, self, _, _ := runtime.Caller(0)
	f, err := os.Open(filepath.Join(filepath.Dir(self), "allow.txt"))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	ids := map[string]bool{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		ids[line] = true
	}

	return ids, s.Err()
}

func main() {
	accept, err := allowed()
	if err != nil {
		fmt.Fprintln(os.Stderr, "reading allow.txt:", err)
		os.Exit(2)
	}

	called := map[string]string{} // osv id -> the function that reaches it
	summary := map[string]string{}

	dec := json.NewDecoder(os.Stdin)
	for {
		var e entry
		if err := dec.Decode(&e); err == io.EOF {
			break
		} else if err != nil {
			fmt.Fprintln(os.Stderr, "parsing govulncheck output:", err)
			os.Exit(2)
		}

		if e.OSV != nil {
			summary[e.OSV.ID] = e.OSV.Summary
		}

		if f := e.Finding; f != nil && len(f.Trace) > 0 && f.Trace[0].Function != "" {
			called[f.OSV] = f.Trace[0].Module + "." + f.Trace[0].Function
		}
	}

	var unexpected, stale []string
	for id := range called {
		if !accept[id] {
			unexpected = append(unexpected, id)
		}
	}

	for id := range accept {
		if _, still := called[id]; !still {
			stale = append(stale, id)
		}
	}

	sort.Strings(unexpected)
	sort.Strings(stale)

	for _, id := range stale {
		fmt.Printf("no longer reachable, drop it from allow.txt: %s\n", id)
	}

	if len(unexpected) == 0 {
		fmt.Printf("%d called, all accepted in allow.txt\n", len(called))
		return
	}

	fmt.Printf("\n%d called and not accepted:\n\n", len(unexpected))
	for _, id := range unexpected {
		fmt.Printf("  %s  %s\n", id, summary[id])
		fmt.Printf("      reached via %s\n", called[id])
		fmt.Printf("      https://pkg.go.dev/vuln/%s\n\n", id)
	}

	fmt.Println("Update the dependency, or add the id to tools/vulncheck/allow.txt with a reason.")
	os.Exit(1)
}
