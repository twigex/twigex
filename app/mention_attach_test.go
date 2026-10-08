// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"testing"

	"github.com/twigex/twigex/model"
)

const otherMentionID = "11111111111111111111111111111111"

func TestAttachMentionUsersNamesEachID(t *testing.T) {
	jane := model.User{ID: testMentionID, Username: "jane", Name: "Jane", LastName: "Doe"}
	a, _ := mentionApp(jane)

	post := model.Post{Message: "hey <@" + testMentionID + ">"}
	a.attachMentionUsers(&post)

	got, ok := post.Mentions[testMentionID]
	if !ok {
		t.Fatalf("no entry for the mentioned id: %v", post.Mentions)
	}
	if got.Username != "jane" || got.Name != "Jane" || got.LastName != "Doe" {
		t.Errorf("got %+v, want jane/Jane/Doe", got)
	}
}

func TestAttachMentionUsersFollowsReplies(t *testing.T) {
	jane := model.User{ID: testMentionID, Username: "jane"}
	bob := model.User{ID: otherMentionID, Username: "bob"}
	a, _ := mentionApp(jane, bob)

	post := model.Post{
		Message: "hey <@" + testMentionID + ">",
		Reply:   &model.Post{Message: "earlier <@" + otherMentionID + ">"},
	}
	a.attachMentionUsers(&post)

	if _, ok := post.Mentions[testMentionID]; !ok {
		t.Errorf("post: no entry for %s", testMentionID)
	}
	if _, ok := post.Reply.Mentions[otherMentionID]; !ok {
		t.Errorf("reply: no entry for %s", otherMentionID)
	}
	if _, ok := post.Mentions[otherMentionID]; ok {
		t.Error("the reply's mention leaked into the parent post")
	}
}

func TestAttachMentionUsersOmitsUnknownID(t *testing.T) {
	a, _ := mentionApp()

	post := model.Post{Message: "hey <@" + testMentionID + ">"}
	a.attachMentionUsers(&post)

	if len(post.Mentions) != 0 {
		t.Errorf("got %v, want no entries", post.Mentions)
	}
}

func TestAttachMentionUsersLeavesPlainPostsAlone(t *testing.T) {
	jane := model.User{ID: testMentionID, Username: "jane"}
	a, _ := mentionApp(jane)

	post := model.Post{Message: "no mentions here, and @jane is the legacy form"}
	a.attachMentionUsers(&post)

	if post.Mentions != nil {
		t.Errorf("got %v, want nil", post.Mentions)
	}
}

func TestAttachMentionUsersHandlesNilPost(t *testing.T) {
	a, _ := mentionApp()
	a.attachMentionUsers(nil)
}

func TestAttachResponseMentionUsersCoversRepliesAndPosts(t *testing.T) {
	jane := model.User{ID: testMentionID, Username: "jane"}
	bob := model.User{ID: otherMentionID, Username: "bob"}
	a, _ := mentionApp(jane, bob)

	res := &model.PostResponse{
		Posts:      []model.Post{{Message: "hey <@" + testMentionID + ">"}},
		ReplyPosts: []model.Post{{Message: "earlier <@" + otherMentionID + ">"}},
	}
	a.attachResponseMentionUsers(res)

	if _, ok := res.Posts[0].Mentions[testMentionID]; !ok {
		t.Errorf("posts: no entry for %s", testMentionID)
	}
	if _, ok := res.ReplyPosts[0].Mentions[otherMentionID]; !ok {
		t.Errorf("reply posts: no entry for %s", otherMentionID)
	}
}

func TestAttachMentionUsersBatchesOneLookup(t *testing.T) {
	jane := model.User{ID: testMentionID, Username: "jane"}
	a, _ := mentionApp(jane)

	calls := 0
	a.Store.User = &countingUserStore{
		fakeMentionUserStore: a.Store.User.(*fakeMentionUserStore),
		calls:                &calls,
	}

	posts := make([]*model.Post, 0, 3)
	for range 3 {
		posts = append(posts, &model.Post{Message: "hey <@" + testMentionID + ">"})
	}

	a.attachMentionUsers(posts...)

	if calls != 1 {
		t.Errorf("got %d lookups, want 1", calls)
	}
	for i, p := range posts {
		if _, ok := p.Mentions[testMentionID]; !ok {
			t.Errorf("post %d: no entry for %s", i, testMentionID)
		}
	}
}

type countingUserStore struct {
	*fakeMentionUserStore
	calls *int
}

func (c *countingUserStore) GetByIDs(ids []string) ([]model.User, error) {
	*c.calls++
	return c.fakeMentionUserStore.GetByIDs(ids)
}
