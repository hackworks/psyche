package utils

import (
	"fmt"
	"regexp"
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

var ignoreQueryTagsFilterRegex, ignoreIndexTagsFilterRegex *regexp.Regexp

func init() {
	// Map of suffix and prefix since we break them at tokenize
	var ignoreQueryWords = []string{
		"@search",
		"@ignore",
		"@silent",
		"@quiet",
		"@find",
		"@register",
		"@botler",
	}

	var ignoreIndexWords []string = append(ignoreQueryWords, []string{
		"@all",
		"@here",
	}...)

	ignoreQueryTagsFilterRegex = regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(ignoreQueryWords, "|")))
	ignoreIndexTagsFilterRegex = regexp.MustCompile(fmt.Sprintf("(%s)", strings.Join(ignoreIndexWords, "|")))
}

func ExtractIndexTags(msg string, pct float64, minWords int, disableHashCheck bool) ([]string, []string) {
	// Check if message is to be ignored
	if ignoreIndexTagsFilterRegex.MatchString(msg) {
		return nil, nil
	}

	// Strip out the ignore words from the query input
	msg = ignoreQueryTagsFilterRegex.ReplaceAllString(msg, "")

	doc := summarize.NewDocument(msg)
	words := tokenize.NewTreebankWordTokenizer().Tokenize(doc.Content)

	var tags, keywords []string
	var prevWord string
	var tagMap = make(map[string]byte)

	for _, w := range words {
		// Store the hash tagMap and @ mentions by ignoring repeats
		if (prevWord == "#" || prevWord == "@") && w != prevWord {
			lw := strings.ToLower(w)
			tags = append(tags, lw)
			tagMap[lw] = 1
		}

		prevWord = w
	}

	// Check if we have sufficient index data to index the message
	if len(tagMap) == 0 && (!disableHashCheck || (minWords > 0 && len(words) < minWords)) {
		return nil, nil
	}

	// Check if we have sufficient keywords with round-off to search this message or enrich it
	moreTags := int(0.5 + (pct*doc.NumWords - float64(len(tagMap))))
	if moreTags > 0 {
		var kw keywordArray
		for k, v := range doc.Keywords() {
			lw := strings.ToLower(k)
			if _, ok := tagMap[lw]; !ok {
				kw = append(kw, keyword{lw, v})
			}
		}
		sort.Sort(kw)

		for _, w := range kw {
			if moreTags == len(keywords) {
				break
			}

			keywords = append(keywords, w.word)
		}
	}

	return tags, keywords
}

// QueryTags parses the search query and returns the operation type and search words
func ExtractQueryTags(msg string) (byte, []string) {
	var queryOp byte

	// Check for query operator
	if strings.ContainsAny(msg, "+&") {
		queryOp = '+'
	}

	// Strip out the ignore words from the query input
	msg = ignoreQueryTagsFilterRegex.ReplaceAllString(msg, "")
	return queryOp, tokenize.NewWordBoundaryTokenizer().Tokenize(msg)
}
