package plugins

import (
	"database/sql"
	"net/url"

	"strings"

	"fmt"

	"bitbucket.org/psyche/types"
	"bitbucket.org/psyche/utils"
	"github.com/lib/pq"
)

type bookmarkPlugin struct {
	db      types.DBH
	plugins Psyches
}

// Minimum number of tags per message
const tagsPerMessage = 0.1

// NewBookmarkPlugin creates an instance of bookmark plugin implementing Psyche interface
func NewBookmarkPlugin(db *sql.DB, p Psyches) Psyche {
	r := &bookmarkPlugin{types.DBH{db}, p}

	// FIXME: DB admin job in the absence of shell access, devise a better approach for one-off jobs
	// r.db.Exec("DROP TABLE bookmark")

	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS bookmark (user_id text, userbase_id text, room_id text, tags text[], keywords text[], ctime timestamp, message text)")
	if err != nil {
		return nil
	}

	return r
}

func (p *bookmarkPlugin) Handle(u *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	if len(rmsg.Message) == 0 {
		return nil, nil
	}

	// Context: userbaseID:chatroomID
	scope := strings.Split(rmsg.Context, ":")
	if len(scope) != 2 {
		return nil, types.ErrBookmark{fmt.Errorf("missing userbase:chatroom for scope")}
	}

	// Extract tags and smart tags from message
	tags, keywords := utils.ExtractTags(rmsg.Message, tagsPerMessage)

	// If we do not have a single tag, this message is not meant for searching
	if len(tags) == 0 {
		return nil, nil
	}

	_, err := p.db.Exec("INSERT INTO bookmark VALUES($1, $2, $3, $4, $5, NOW(), $6)",
		rmsg.Sender.ID, scope[0], scope[1], pq.Array(tags), pq.Array(keywords), rmsg.Message)

	return nil, err
}

func (p *bookmarkPlugin) Refresh() error {
	return nil
}
