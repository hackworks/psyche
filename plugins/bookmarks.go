package plugins

import (
	"database/sql"
	"net/url"
	"strings"

	"bitbucket.org/psyche/types"
	"github.com/jdkato/prose/tokenize"
	"github.com/lib/pq"
)

type bookmarkPlugin struct {
	db      types.DBH
	plugins Psyches
}

// NewBookmarkPlugin creates an instance of bookmark plugin implementing Psyche interface
func NewBookmarkPlugin(db *sql.DB, p Psyches) Psyche {
	r := &bookmarkPlugin{types.DBH{db}, p}

	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS bookmarks (user_id text, room_id text, tags text[], ctime date, message text)")
	if err != nil {
		return nil
	}

	return r
}

func (p *bookmarkPlugin) Handle(u *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	words := tokenize.NewTreebankWordTokenizer().Tokenize(rmsg.Message)

	var tags []string
	var isTag bool
	for _, w := range words {
		if w == "#" {
			isTag = true
			continue
		}

		if isTag {
			isTag = false
			tags = append(tags, strings.ToLower(w))
		}
	}

	// If there are no tags, bail out
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
