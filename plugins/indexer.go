package plugins

import (
	"database/sql"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"bitbucket.org/psyche/types"
	"bitbucket.org/psyche/utils"
	"github.com/lib/pq"
)

type indexerPlugin struct {
	db      types.DBH
	plugins Psyches
}

// Minimum number of tags per message
const tagsPerMessage = 0.1

// Minimum number of words in a message without tags
const minWordsPerMessage = 5

// NewIndexerPlugin creates an instance of indexer plugin implementing Psyche interface
func NewIndexerPlugin(db *sql.DB, p Psyches) Psyche {
	r := &indexerPlugin{types.DBH{db}, p}

	// FIXME: DB admin job in the absence of shell access, devise a better approach for one-off jobs
	// r.db.Exec("DROP TABLE indexer")

	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS indexer (user_id text, userbase_id text, room_id text, tags text[], keywords text[], ctime timestamp, message text)")
	if err != nil {
		return nil
	}

	return r
}

func (p *indexerPlugin) Handle(u *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	if len(rmsg.Message) == 0 ||
		// Explicitly ignore messages from botler
		rmsg.Sender.ID == "557058:48faede9-ea1d-4bf0-8a33-07d02c1fe6c6" ||
		rmsg.Sender.ID == "557058:58827303-cf25-4168-8846-6ae6080b1993" {
		return nil, nil
	}

	disableHashCheck, _ := strconv.ParseBool(u.Query().Get("disableHashCheck"))

	// Context: userbaseID:chatroomID
	scope := strings.SplitN(rmsg.Context, ":", 2)
	if len(scope) != 2 {
		return nil, types.ErrIndexer{fmt.Errorf("missing userbase:chatroom for scope")}
	}

	// Extract tags and smart tags from message
	tags, keywords := utils.ExtractIndexTags(rmsg.Message, tagsPerMessage, minWordsPerMessage, disableHashCheck)

	if !disableHashCheck && len(tags) == 0 {
		return nil, nil
	}

	_, err := p.db.Exec("INSERT INTO indexer VALUES($1, $2, $3, $4, $5, NOW(), $6)",
		rmsg.Sender.ID, scope[0], scope[1], pq.Array(tags), pq.Array(keywords), rmsg.Message)

	return nil, err
}

func (p *indexerPlugin) Refresh() error {
	return nil
}
