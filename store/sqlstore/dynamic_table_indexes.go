// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

// dynamicTableLayout is the index set a per-workspace table of one kind is
// created with. unique names the indexes, by their comma-joined columns, that
// no two rows may share.
type dynamicTableLayout struct {
	indexes [][]string
	unique  []string
}

var (
	taskTableLayout = dynamicTableLayout{
		// The tasks that head the grid's branches are counted from
		// (parent_task_id, deleted_at) alone; it serves lookups by parent too.
		indexes: [][]string{{"assignee"}, {"status"}, {"start_date"}, {"due_date"}, {"created_at"}, {"updated_at"}, {"created_by"}, {"parent_task_id", "deleted_at"}},
	}
	// A task links to another task at most once in a link field.
	linkTableLayout = dynamicTableLayout{
		indexes: [][]string{{"table_item_id", "parent_table_item_id"}, {"parent_table_item_id"}},
		unique:  []string{"table_item_id,parent_table_item_id"},
	}
	selectTableLayout = dynamicTableLayout{
		indexes: [][]string{{"name"}, {"color"}},
	}
)
