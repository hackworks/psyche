package plugins

import (
	"database/sql"
	"net/url"
	"sort"
	"strings"

	"bitbucket.org/psyche/types"
	"github.com/jdkato/prose/summarize"
	"github.com/jdkato/prose/tokenize"
	"github.com/lib/pq"
)

type bookmarkPlugin struct {
	db      types.DBH
	plugins Psyches
}

// Map of suffix and prefix since we break them at tokenize
var ignoreFilter = map[string]string{
	"search": "@",
	"ignore": "@",
	"silent": "@",
	"quiet":  "@",
}

// Minimum number of tags per message
const tagsPerMessage = 0.05

// NewBookmarkPlugin creates an instance of bookmark plugin implementing Psyche interface
func NewBookmarkPlugin(db *sql.DB, p Psyches) Psyche {
	r := &bookmarkPlugin{types.DBH{db}, p}

	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS bookmarks (user_id text, room_id text, tags text[], ctime date, message text)")
	if err != nil {
		return nil
	}

	return r
}

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

func (p *bookmarkPlugin) Handle(u *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	if len(rmsg.Message) == 0 {
		return nil, nil
	}

	// Extract tags and smart tags from message
	tags := p.extractTags(rmsg.Message, tagsPerMessage)

	// If we do not have a single tag, this message is not meant for searching
	if len(tags) == 0 {
		return nil, nil
	}

	_, err := p.db.Exec("INSERT INTO bookmarks VALUES($1, $2, $3, NOW(), $4)",
		rmsg.Sender.ID, rmsg.Context, pq.Array(tags), rmsg.Message)

	return nil, err
}

func (p *bookmarkPlugin) Refresh() error {
	return nil
}

func (p *bookmarkPlugin) extractTags(msg string, pct float64) []string {
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

	// Check if we have sufficient keywords to search this message or enrich it
	moreTags := pct*doc.NumWords - float64(len(tags))
	if moreTags > 0 {
		var kw keywordArray
		for k, v := range doc.Keywords() {
			kw = append(kw, keyword{k, v})
		}
		sort.Sort(kw)

		for cc := 0; cc < int(moreTags); cc++ {
			tags = append(tags, kw[cc].word)
		}
	}

	return tags
}
