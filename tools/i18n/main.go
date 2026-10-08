// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type i18nEntry struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

// scanDirs are the packages where translation IDs are referenced, either as the
// first argument to model.NewAppError or as the message ID passed to i18n.T/i18n.Tf.
var scanDirs = []string{"app", "api"}

// overlayDirs returns the other modules in the Go workspace, which is how a
// private build carries packages that are not in this repository. Reading the
// workspace rather than naming a module keeps that arrangement out of here, and
// means a second overlay needs no change. It returns nothing when there is no
// workspace, which is the case for a plain clone and in CI.
func overlayDirs() []string {
	out, err := exec.Command("go", "work", "edit", "-json").Output()
	if err != nil {
		return nil
	}

	var work struct {
		Use []struct{ DiskPath string }
	}
	if err := json.Unmarshal(out, &work); err != nil {
		return nil
	}

	var dirs []string
	for _, u := range work.Use {
		if u.DiskPath == "." || u.DiskPath == "" {
			continue
		}

		if _, err := os.Stat(u.DiskPath); err != nil {
			continue
		}

		dirs = append(dirs, u.DiskPath)
	}

	return dirs
}

// scan returns the IDs referenced as string literals (mapped to the files using
// them) and the set of every string literal seen, used to resolve IDs built by
// concatenation such as prefix + ".subject".
func scan(dirs []string) (map[string][]string, map[string]bool, error) {
	used := make(map[string][]string)
	literals := make(map[string]bool)
	fset := token.NewFileSet()

	for _, dir := range dirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() || !strings.HasSuffix(path, ".go") {
				return nil
			}

			f, err := parser.ParseFile(fset, path, nil, 0)
			if err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}

			ast.Inspect(f, func(n ast.Node) bool {
				if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					if s, err := strconv.Unquote(lit.Value); err == nil {
						literals[s] = true
					}

					return true
				}

				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkg, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}

				argIdx := -1
				switch {
				case pkg.Name == "model" && sel.Sel.Name == "NewAppError":
					argIdx = 0
				case pkg.Name == "i18n" && (sel.Sel.Name == "T" || sel.Sel.Name == "Tf"):
					argIdx = 1
				}

				if argIdx < 0 || len(call.Args) <= argIdx {
					return true
				}

				lit, ok := call.Args[argIdx].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}

				if s, err := strconv.Unquote(lit.Value); err == nil {
					used[s] = append(used[s], path)
				}

				return true
			})

			return nil
		})
		if err != nil {
			return nil, nil, err
		}
	}

	return used, literals, nil
}

var tmplCallRe = regexp.MustCompile(`\bT\s+"([^"]+)"`)

// scanTemplates records IDs referenced through the {{T "id"}} template function.
func scanTemplates(dir string, used map[string][]string) error {
	return filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		for _, m := range tmplCallRe.FindAllStringSubmatch(string(b), -1) {
			used[m[1]] = append(used[m[1]], path)
		}

		return nil
	})
}

func main() {
	data, err := os.ReadFile("i18n/en.json")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read i18n/en.json: %v\n", err)
		os.Exit(1)
	}

	var entries []i18nEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		fmt.Fprintf(os.Stderr, "error: parse i18n/en.json: %v\n", err)
		os.Exit(1)
	}

	defined := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		defined[e.ID] = struct{}{}
	}

	overlays := overlayDirs()
	dirs := append(append([]string{}, scanDirs...), overlays...)

	used, literals, err := scan(dirs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: scan: %v\n", err)
		os.Exit(1)
	}

	if err := scanTemplates("templates", used); err != nil {
		fmt.Fprintf(os.Stderr, "error: scan templates: %v\n", err)
		os.Exit(1)
	}

	// An ID counts as used if it appears as a literal, or if it splits into a
	// literal prefix and a literal suffix (covers prefix + ".subject" patterns).
	isUsed := func(id string) bool {
		if _, ok := used[id]; ok {
			return true
		}

		if literals[id] {
			return true
		}

		for l := range literals {
			if l != "" && l != id && strings.HasPrefix(id, l) && literals[id[len(l):]] {
				return true
			}
		}

		return false
	}

	var missing []string
	for id := range used {
		if _, ok := defined[id]; !ok {
			missing = append(missing, id)
		}
	}

	// An ID used only by an overlay module looks unused without it, so the report
	// is withheld rather than listing false positives.
	var unused []string
	if len(overlays) > 0 {
		for id := range defined {
			if !isUsed(id) {
				unused = append(unused, id)
			}
		}
	}

	if len(missing) > 0 {
		fmt.Println("MISSING from i18n/en.json (used in app but not translated):")
		for _, id := range missing {
			fmt.Printf("  %s  (in: %s)\n", id, strings.Join(used[id], ", "))
		}
	}

	if len(unused) > 0 {
		fmt.Println("UNUSED in app (defined in i18n/en.json but never used):")
		for _, id := range unused {
			fmt.Printf("  %s\n", id)
		}
	}

	if len(overlays) == 0 {
		fmt.Printf("Checked %s. No workspace overlay, so unused ids are not reported.\n",
			strings.Join(scanDirs, ", "))
	}

	if len(missing) == 0 && len(unused) == 0 {
		fmt.Println("All translation IDs are consistent.")
		return
	}

	fmt.Printf("\n%d missing, %d unused.\n", len(missing), len(unused))

	if len(missing) > 0 {
		os.Exit(1)
	}
}
