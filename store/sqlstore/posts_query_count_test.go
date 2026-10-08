// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package sqlstore

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

// Capped at one connection so SHOW SESSION STATUS reports the counter for the
// session the queries ran on; the shared testDB pools and would report
// whichever connection the status query landed on.
func pinnedConn(t *testing.T) *sql.DB {
	t.Helper()

	cfg, err := mysql.ParseDSN(os.Getenv("TEST_MYSQL_DSN"))
	if err != nil {
		t.Skipf("%v", err)
	}
	cfg.DBName = testDBName
	cfg.MultiStatements = false

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { db.Close() })

	return db
}

func queriesRun(t *testing.T, db *sql.DB) int {
	t.Helper()

	var name string
	var n int
	if err := db.QueryRow("SHOW SESSION STATUS LIKE 'Questions'").Scan(&name, &n); err != nil {
		t.Fatalf("read session status: %v", err)
	}

	return n
}

// The read used to go back to the database for every post's reply, attachments
// and reactions, so thirty messages cost upwards of a hundred round trips.
// Counting rather than timing states the property instead of measuring the
// machine the test runs on.
func TestChannelReadQueryCountDoesNotGrowWithPosts(t *testing.T) {
	requireDB(t)
	db := pinnedConn(t)
	repo := &postRepository{Db: db}

	count := func(channel string, posts int) int {
		cleanTables(t, "posts", "post_reactions", "post_attachments")

		base := time.Now().UnixMilli()
		root := seedPostAt(t, repo, channel, "user-1", "", "root", base)
		seedAttachment(t, repo, channel, root, "user-1")
		for i := 1; i < posts; i++ {
			id := seedPostAt(t, repo, channel, "user-1", root, "m", base+int64(i))
			seedAttachment(t, repo, channel, id, "user-1")
			seedReaction(t, repo, channel, id, "user-1", ":+1:")
		}

		before := queriesRun(t, db)
		got, err := repo.GetAllForChannel(t.Context(), channel, "", "")
		if err != nil {
			t.Fatalf("GetAllForChannel: %v", err)
		}
		if len(got.Posts) != posts {
			t.Fatalf("read %d posts, seeded %d", len(got.Posts), posts)
		}
		after := queriesRun(t, db)

		// Both status reads are themselves queries; the second is counted.
		return after - before - 1
	}

	small := count("channel-small", 5)
	large := count("channel-large", 30)

	if small != large {
		t.Errorf("5 posts cost %d queries, 30 posts cost %d: the count grows with the number of posts",
			small, large)
	}
	if large > 6 {
		t.Errorf("a channel read costs %d queries, want the posts plus a batch each for replies, attachments and reactions", large)
	}
	t.Logf("channel read costs %d queries regardless of post count", large)
}
