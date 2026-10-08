// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

// Syncs all translation JSON files against their en.json master.
// Run from the repo root: go run ./tools/sync-translations
//
// Backend  (i18n/):              array of {id, translation}, missing entries get translation: ""
// Frontend (frontend/src/i18n/): flat {key: value}, missing entries get value: ""
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type backendEntry struct {
	ID          string `json:"id"`
	Translation string `json:"translation"`
}

func syncBackend(dir string) error {
	masterPath := filepath.Join(dir, "en.json")
	masterData, err := os.ReadFile(masterPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", masterPath, err)
	}

	var master []backendEntry
	if err := json.Unmarshal(masterData, &master); err != nil {
		return fmt.Errorf("parse %s: %w", masterPath, err)
	}

	masterIDs := make([]string, len(master))
	masterSet := make(map[string]struct{}, len(master))
	for i, e := range master {
		masterIDs[i] = e.ID
		masterSet[e.ID] = struct{}{}
	}

	return walkJSONFiles(dir, func(name, filePath string) error {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read %s: %w", filePath, err)
		}

		var existing []backendEntry
		if len(bytes.TrimSpace(data)) > 0 {
			if err := json.Unmarshal(data, &existing); err != nil {
				return fmt.Errorf("parse %s: %w", filePath, err)
			}
		}

		lookup := make(map[string]string, len(existing))
		for _, e := range existing {
			lookup[e.ID] = e.Translation
		}

		changed := false
		result := make([]backendEntry, 0, len(masterIDs))

		for _, id := range masterIDs {
			if val, ok := lookup[id]; ok {
				result = append(result, backendEntry{ID: id, Translation: val})
			} else {
				result = append(result, backendEntry{ID: id, Translation: ""})
				fmt.Printf("  + added   [%s] %s\n", name, id)
				changed = true
			}
		}

		for _, e := range existing {
			if _, ok := masterSet[e.ID]; !ok {
				fmt.Printf("  - removed [%s] %s\n", name, e.ID)
				changed = true
			}
		}

		if !changed {
			return nil
		}

		out, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal %s: %w", filePath, err)
		}

		if err := os.WriteFile(filePath, append(out, '\n'), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", filePath, err)
		}

		fmt.Printf("  saved %s\n", filePath)
		return nil
	})
}

func syncFrontend(dir string) error {
	masterPath := filepath.Join(dir, "en.json")
	masterData, err := os.ReadFile(masterPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", masterPath, err)
	}

	masterKeys, masterSet, err := orderedKeys(masterData)
	if err != nil {
		return fmt.Errorf("parse %s: %w", masterPath, err)
	}

	return walkJSONFiles(dir, func(name, filePath string) error {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read %s: %w", filePath, err)
		}

		existing := make(map[string]string)
		if len(bytes.TrimSpace(data)) > 0 {
			if err := json.Unmarshal(data, &existing); err != nil {
				return fmt.Errorf("parse %s: %w", filePath, err)
			}
		}

		changed := false
		keys := make([]string, 0, len(masterKeys))
		vals := make([]string, 0, len(masterKeys))

		for _, key := range masterKeys {
			keys = append(keys, key)
			if val, ok := existing[key]; ok {
				vals = append(vals, val)
			} else {
				vals = append(vals, "")
				fmt.Printf("  + added   [%s] %s\n", name, key)
				changed = true
			}
		}

		for key := range existing {
			if _, ok := masterSet[key]; !ok {
				fmt.Printf("  - removed [%s] %s\n", name, key)
				changed = true
			}
		}

		if !changed {
			return nil
		}

		out, err := marshalFlatJSON(keys, vals)
		if err != nil {
			return fmt.Errorf("marshal %s: %w", filePath, err)
		}

		if err := os.WriteFile(filePath, out, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", filePath, err)
		}

		fmt.Printf("  saved %s\n", filePath)
		return nil
	})
}

// walkJSONFiles calls fn for every *.json file in dir except en.json.
func walkJSONFiles(dir string, fn func(name, path string) error) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read dir %s: %w", dir, err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") || e.Name() == "en.json" {
			continue
		}

		if err := fn(e.Name(), filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}

	return nil
}

// orderedKeys extracts JSON object keys in their declaration order.
func orderedKeys(data []byte) (keys []string, set map[string]struct{}, err error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	tok, err := dec.Token()
	if err != nil {
		return nil, nil, err
	}

	if tok != json.Delim('{') {
		return nil, nil, fmt.Errorf("expected '{', got %v", tok)
	}

	set = make(map[string]struct{})
	for dec.More() {
		tok, err = dec.Token()
		if err != nil {
			return nil, nil, err
		}

		key := tok.(string)
		keys = append(keys, key)
		set[key] = struct{}{}

		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return nil, nil, err
		}
	}

	return keys, set, nil
}

// json.Marshal escapes <, > and & as < and friends, which rewrites every
// translation containing markup each time this runs. An encoder can be told not
// to, at the cost of a trailing newline to trim.
func marshalString(s string) ([]byte, error) {
	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return nil, err
	}

	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// marshalFlatJSON writes {"key": "value"} with 4-space indentation, preserving key order.
func marshalFlatJSON(keys, vals []string) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, key := range keys {
		kb, err := marshalString(key)
		if err != nil {
			return nil, err
		}

		vb, err := marshalString(vals[i])
		if err != nil {
			return nil, err
		}

		buf.WriteString("    ")
		buf.Write(kb)
		buf.WriteString(": ")
		buf.Write(vb)
		if i < len(keys)-1 {
			buf.WriteByte(',')
		}

		buf.WriteByte('\n')
	}

	buf.WriteString("}\n")
	return buf.Bytes(), nil
}

func main() {
	fmt.Println("\nbackend: i18n")
	if err := syncBackend("i18n"); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nfrontend: frontend/src/i18n")
	if err := syncFrontend("frontend/src/i18n"); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nDone.")
}
