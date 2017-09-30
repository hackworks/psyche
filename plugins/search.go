package plugins

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"bitbucket.org/psyche/types"
	"github.com/jdkato/prose/summarize"
	"github.com/lib/pq"
)

type searchPlugin struct {
	db      types.DBH
	plugins Psyches
}

// Limit the number of search results to prevent clogging output
const resultLimit = 100

// NewBookmarkPlugin creates an instance of bookmark plugin implementing Psyche interface
func NewSearchPlugin(db *sql.DB, p Psyches) Psyche {
	return &searchPlugin{types.DBH{db}, p}
}

func (p *searchPlugin) Handle(url *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	val, ok := p.plugins["relay"]
	if !ok {
		return nil, types.ErrSearch{errors.New("failed to get replay plugin")}
	}

	relay, ok := val.(*relayPlugin)
	if !ok {
		return nil, types.ErrSearch{errors.New("failed to cast replay plugin interface")}
	}

	target := url.Query().Get("target")
	if len(target) == 0 {
		return nil, types.ErrSearch{errors.New("target room to send results missing")}
	}

	doc := summarize.NewDocument(rmsg.Message)

	// If there are no tags, bail out
	if doc.NumWords == 0 {
		return nil, nil
	}

	// Since we store tags without '#', strip them is someone puts them in
	var tags []string
	for w, _ := range doc.WordFrequency {
		tags = append(tags, strings.ToLower(w))
	}

	// TODO:
	// * Well defined search syntax
	// * Date range based search
	// * Support basic AND operation
	// * Search for self tagged messages across rooms
	// * Suggest tags to limit search
	// * Background search jobs for more heuristics in the future

	var err error
	var rows *sql.Rows
	scope := url.Query().Get("scope")

	switch scope {
	case "self", "me", "mine", "myself":
		rows, err = p.db.Query("SELECT TO_CHAR(ctime, 'MM-DD-YYYY'), message FROM bookmarks WHERE room_id=$1 AND $2 && tags AND user_id=$4 ORDER BY ctime DESC LIMIT $3",
			rmsg.Context, pq.Array(tags), resultLimit+1, rmsg.Sender.ID)
	case "room", "chatroom", "conversation":
		fallthrough
	default:
		rows, err = p.db.Query("SELECT TO_CHAR(ctime, 'MM-DD-YYYY'), message FROM bookmarks WHERE room_id=$1 AND $2 && tags ORDER BY ctime DESC LIMIT $3",
			rmsg.Context, pq.Array(tags), resultLimit+1)
	}

	if err != nil {
		return nil, err
	}

	var resultCount int
	var msg, ct string
	var buff bytes.Buffer
	for rows.Next() {
		err = rows.Scan(&ct, &msg)
		if err != nil {
			break
		}

		resultCount++

		// NOTE: We fetch 1 more than the limit to determine if there are more results than the limit
		if resultCount < resultLimit {
			buff.WriteString(fmt.Sprintf("\n%s >\n%s\n", ct, msg))
		}
	}

	if resultCount > 0 {
		// Provide hints if results are truncated due to search limit
		if resultCount > resultLimit {
			buff.WriteString(fmt.Sprintf("showing %d results, try refining search:\n", resultLimit))
		} else {
			buff.WriteString(fmt.Sprintf("showing %d results:\n", resultCount))
		}

		smsg := types.SendMsg{buff.String(), "text"}
		relay.RelayMsg(target, &smsg)
	}

	return nil, err
}

func (p *searchPlugin) Refresh() error {
	return nil
}
