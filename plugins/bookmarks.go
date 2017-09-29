package plugins

import (
	"database/sql"
	"net/url"

	"bitbucket.org/psyche/types"
)

type bookmarkPlugin struct {
	db      types.DBH
	plugins Psyches
}

// NewBookmarkPlugin creates an instance of bookmark plugin implementing Psyche interface
func NewBookmarkPlugin(db *sql.DB, p Psyches) Psyche {
	r := &bookmarkPlugin{types.DBH{db}, p}

	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS bookmarks (user_id text, tags text[], ctime timestamp, message text, PRIMARY KEY (user_id))")
	if err != nil {
		return nil
	}

	return r
}

func (p *bookmarkPlugin) Handle(*url.URL, *types.RecvMsg) (*types.SendMsg, error) {
	return nil, nil
}

func (p *bookmarkPlugin) Refresh() error {
	return nil
}
