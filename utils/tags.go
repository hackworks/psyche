package utils

import (
	"sort"
	"strings"

	"github.com/jdkato/prose/summarize"
	"github.com/jdkato/prose/tokenize"
)

// Custom type to sort keywords in a message based on frequence
type keyword struct {
	word string
	freq int
}
type keywordArray []keyword

func (s keywordArray) Len() int {
	return len(s)
}

func (s keywordArray) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s keywordArray) Less(i, j int) bool {
	return s[i].freq > s[j].freq
}

// Map of suffix and prefix since we break them at tokenize
var ignoreFilter = map[string]string{
	"search": "@",
	"ignore": "@",
	"silent": "@",
	"quiet":  "@",
}

func ExtractTags(msg string, pct float64) []string {
	doc := summarize.NewDocument(msg)
	words := tokenize.NewTreebankWordTokenizer().Tokenize(doc.Content)

	var tags []string
	var prevWord string
	for _, w := range words {
		// Look for ignore filter and skip bookmarking them
		if p, ok := ignoreFilter[w]; ok && prevWord == p {
			return nil
		}

		// Store the hash tags and ignore multiple hashes
		if prevWord == "#" && w != "#" {
			tags = append(tags, strings.ToLower(w))
		}

		prevWord = w
	}

	// TODO: For now, let us index only messages with explicit tags words
	if len(tags) == 0 {
		return nil
	}

	// Check if we have sufficient keywords with round-off to search this message or enrich it
	moreTags := int(0.5 + (pct*doc.NumWords - float64(len(tags))))
	if moreTags > 0 {
		var kw keywordArray
		for k, v := range doc.Keywords() {
			kw = append(kw, keyword{k, v})
		}
		sort.Sort(kw)

		for cc := 0; cc < moreTags; cc++ {
			tags = append(tags, kw[cc].word)
		}
	}

	return tags
}
