// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// surroundingModule is the module of the repository the engine is in. The
// engine must use none of its packages, so it can become a project of its own.
const surroundingModule = "github.com/twigex/"

func TestEngineUsesNoPackageOfTheSurroundingModule(t *testing.T) {
	fileNames, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("list files: %v", err)
	}

	files := token.NewFileSet()

	for _, fileName := range fileNames {
		source, err := os.ReadFile(fileName)
		if err != nil {
			t.Fatalf("read %s: %v", fileName, err)
		}

		parsed, err := parser.ParseFile(files, fileName, source, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("parse %s: %v", fileName, err)
		}

		for _, imported := range parsed.Imports {
			path, err := strconv.Unquote(imported.Path.Value)
			if err != nil {
				t.Fatalf("read import in %s: %v", fileName, err)
			}

			if strings.HasPrefix(path, surroundingModule) {
				t.Errorf("%s uses %s", fileName, path)
			}
		}
	}
}
