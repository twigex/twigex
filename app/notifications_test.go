// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"reflect"
	"testing"

	"github.com/twigex/twigex/model"
)

// Mirrors frontend/src/utils/__tests__/mentions.spec.js, which runs the same
// corpus through splitMentionTokens. Keep the two in step.
const testMentionID = "0f8fad5bd9cb469fa16570867728950e"

func handlesOf(mentions []mention) []string {
	out := make([]string, 0, len(mentions))
	for _, m := range mentions {
		out = append(out, m.Handle)
	}
	return out
}

func TestExtractMentionsPlainHandles(t *testing.T) {
	got := handlesOf(extractMentions("hey @bob and @all"))
	if want := []string{"bob", "all"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExtractMentionsDirectoryStyleNames(t *testing.T) {
	got := handlesOf(extractMentions("ping @jane.doe and @jane-doe and @jane_doe"))
	if want := []string{"jane.doe", "jane-doe", "jane_doe"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// Sentence punctuation sits directly against the handle far more often than it
// forms part of one, so a trailing dot or hyphen belongs to the prose.
func TestExtractMentionsDropsTrailingPunctuation(t *testing.T) {
	got := handlesOf(extractMentions("ask @jane.doe. then @bob- or @carol_"))
	if want := []string{"jane.doe", "bob", "carol"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExtractMentionsIgnoresCodeAndLinks(t *testing.T) {
	got := extractMentions("`@bob` and ```\n@carol\n``` and [@dave](https://example.com)")
	if len(got) != 0 {
		t.Errorf("got %v, want none", got)
	}
}

func TestExtractMentionsDeduplicates(t *testing.T) {
	got := handlesOf(extractMentions("@jane.doe @jane.doe @jane.doe"))
	if want := []string{"jane.doe"}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// Anything the renderer treats as a handle must be storable, or a mention could
// name a user that could never exist.
func TestExtractMentionsOnlyYieldsStorableUsernames(t *testing.T) {
	message := "@jane.doe @bob- @carol_ @1jane @a @jane..doe"

	for _, m := range extractMentions(message) {
		if m.Handle == "all" || m.Handle == "here" {
			continue
		}
		if !model.ValidUsername(model.NormalizeUsername(m.Handle)) {
			t.Errorf("handle %q is not a storable username", m.Handle)
		}
	}
}

func TestExtractMentionsIDForm(t *testing.T) {
	got := extractMentions("hey <@" + testMentionID + "> there")
	if want := []mention{{ID: testMentionID}}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExtractMentionsMixedForms(t *testing.T) {
	got := extractMentions("<@" + testMentionID + "> and @bob and @all")
	want := []mention{
		{ID: testMentionID},
		{Handle: "bob"},
		{Handle: "all"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExtractMentionsMalformedIDFallsBackToName(t *testing.T) {
	got := extractMentions("<@abc123> and <@" + testMentionID + "x>")
	want := []mention{
		{Handle: "abc123"},
		{Handle: testMentionID + "x"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestExtractMentionsIgnoresIDsInCodeAndLinks(t *testing.T) {
	id := testMentionID
	got := extractMentions("`<@" + id + ">` and ```\n<@" + id + ">\n``` and [<@" + id + ">](https://example.com)")
	if len(got) != 0 {
		t.Errorf("got %v, want none", got)
	}
}

// Collapsing the two to one user is resolveHandles' job, not the parser's.
func TestExtractMentionsDeduplicatesAcrossForms(t *testing.T) {
	got := extractMentions("<@" + testMentionID + "> <@" + testMentionID + "> @bob @bob")
	want := []mention{
		{ID: testMentionID},
		{Handle: "bob"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// The id pattern is pinned to the shape model.NewID() produces. If id
// generation ever changes, every stored mention stops parsing, so fail here
// rather than silently in the renderer.
func TestMentionIDPatternMatchesGeneratedIDs(t *testing.T) {
	for range 100 {
		id := model.NewID()
		got := extractMentions("<@" + id + ">")
		if want := []mention{{ID: id}}; !reflect.DeepEqual(got, want) {
			t.Fatalf("generated id %q did not parse as a mention: got %v", id, got)
		}
	}
}

func TestAddedMentionsIgnoresExisting(t *testing.T) {
	got := addedMentions("hey @bob", "hey @bob and @carol")
	if want := []mention{{Handle: "carol"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestAddedMentionsIgnoresExistingIDForm(t *testing.T) {
	previous := "hey <@" + testMentionID + ">"
	got := addedMentions(previous, previous+" and @carol")
	if want := []mention{{Handle: "carol"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
