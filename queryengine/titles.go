// Copyright (c) 2022 Twigex, SIA.
// SPDX-License-Identifier: AGPL-3.0-only

package queryengine

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// These split a name written in capitals into words: "orderDate" becomes
// "order Date", and "HTTPServer" becomes "HTTP Server". A plural of capitals
// stays one word, so "userIDs" gives "User IDs", not "User I Ds".
var (
	pluralCapitals     = regexp.MustCompile(`([A-Z]{2,})s([A-Z_]|$)`)
	capitalsBeforeWord = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
	lowerBeforeCapital = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

var wordsKeptLowercase = map[string]bool{
	"a":    true,
	"an":   true,
	"and":  true,
	"at":   true,
	"but":  true,
	"by":   true,
	"for":  true,
	"from": true,
	"in":   true,
	"into": true,
	"nor":  true,
	"of":   true,
	"off":  true,
	"on":   true,
	"onto": true,
	"or":   true,
	"out":  true,
	"over": true,
	"so":   true,
	"the":  true,
	"to":   true,
	"with": true,
}

// MakeTitle turns a name such as "created_at" or "createdAt" into a title such
// as "Created at".
func MakeTitle(name string) string {
	words := splitIntoWords(name)

	titleWords := make([]string, 0, len(words))
	for position, word := range words {
		isFirstWord := position == 0
		titleWord := titleForWord(word, isFirstWord)

		titleWords = append(titleWords, titleWord)
	}

	return strings.Join(titleWords, " ")
}

func splitIntoWords(name string) []string {
	separated := keepPluralCapitalsTogether(name)
	separated = capitalsBeforeWord.ReplaceAllString(separated, "${1}_${2}")
	separated = lowerBeforeCapital.ReplaceAllString(separated, "${1}_${2}")
	separated = strings.ToLower(separated)

	parts := strings.Split(separated, "_")

	words := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}

		words = append(words, part)
	}

	return words
}

// keepPluralCapitalsTogether writes "IDs" as "IDS_", one word ending there. It
// repeats until nothing changes, because a match takes the capital after it,
// so "IDsABs" needs two rounds. Each round leaves one lowercase s fewer.
func keepPluralCapitalsTogether(name string) string {
	for {
		changed := pluralCapitals.ReplaceAllString(name, "${1}S_${2}")
		if changed == name {
			return name
		}

		name = changed
	}
}

func titleForWord(word string, isFirstWord bool) string {
	if word == "id" {
		return "ID"
	}

	if word == "ids" {
		return "IDs"
	}

	if !isFirstWord && wordsKeptLowercase[word] {
		return word
	}

	return capitalizeFirstLetter(word)
}

func capitalizeFirstLetter(word string) string {
	firstLetter, size := utf8.DecodeRuneInString(word)
	rest := word[size:]

	return string(unicode.ToUpper(firstLetter)) + rest
}
