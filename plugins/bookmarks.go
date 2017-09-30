package plugins

import (
	"database/sql"
	"net/url"

	"bitbucket.org/psyche/types"
	"bitbucket.org/psyche/utils"
	"github.com/lib/pq"
)

type bookmarkPlugin struct {
	db      types.DBH
	plugins Psyches
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

func (p *bookmarkPlugin) Handle(u *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	if len(rmsg.Message) == 0 {
		return nil, nil
	}

	// Extract tags and smart tags from message
	tags := utils.ExtractTags(rmsg.Message, tagsPerMessage)

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
