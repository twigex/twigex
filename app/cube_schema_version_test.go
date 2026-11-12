// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"testing"

	"github.com/twigex/twigex/model"
)

func schemaFiles() []model.CollimatoWorkspaceFile {
	return []model.CollimatoWorkspaceFile{
		{Name: "orders.yml", FileType: model.CollimatoFileTypeCube, Content: "cubes: [orders]", UpdatedAt: 100},
		{Name: "customers.yml", FileType: model.CollimatoFileTypeCube, Content: "cubes: [customers]", UpdatedAt: 100},
		{Name: "sales.yml", FileType: model.CollimatoFileTypeView, Content: "views: [sales]", UpdatedAt: 100},
	}
}

func TestSchemaFingerprintIgnoresOrder(t *testing.T) {
	files := schemaFiles()
	reversed := []model.CollimatoWorkspaceFile{files[2], files[1], files[0]}

	if schemaFingerprint(files) != schemaFingerprint(reversed) {
		t.Error("want the same fingerprint for the same files in any order")
	}
}

func TestSchemaFingerprintChangesOnEveryChange(t *testing.T) {
	original := schemaFingerprint(schemaFiles())

	edited := schemaFiles()
	edited[0].Content = "cubes: [orders, refunds]"

	deleted := schemaFiles()[1:]

	renamed := schemaFiles()
	renamed[2].Name = "revenue.yml"

	cases := map[string][]model.CollimatoWorkspaceFile{
		"an edit in the same second": edited,
		"a deleted file":             deleted,
		"a renamed file":             renamed,
	}

	for name, files := range cases {
		t.Run(name, func(t *testing.T) {
			if schemaFingerprint(files) == original {
				t.Error("want a new fingerprint")
			}
		})
	}
}
