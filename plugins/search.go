package plugins

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"bitbucket.org/psyche/types"
	"bitbucket.org/psyche/utils"
	"github.com/lib/pq"
)

type searchPlugin struct {
	db      types.DBH
	plugins Psyches
}

// Limit the number of search results to prevent clogging output
const resultLimit = 50

// NewBookmarkPlugin creates an instance of bookmark plugin implementing Psyche interface
func NewSearchPlugin(db *sql.DB, p Psyches) Psyche {
	return &searchPlugin{types.DBH{db}, p}
}

func (p *searchPlugin) Handle(url *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	// Context: userbaseID:chatroomID
	scope := strings.Split(rmsg.Context, ":")
	if len(scope) != 2 {
		return nil, types.ErrSearch{fmt.Errorf("missing userbase:chatroom for scope")}
	}

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
		// Look for user registered room for sending messages (UserbaseId:AAID)
		target = scope[0] + ":" + rmsg.Sender.ID
	}

	// TODO:
	// * Well defined search syntax
	// * Date range based search
	// * Search for self tagged messages across rooms
	// * Suggest tags to limit search
	// * Background search jobs for more heuristics in the future

	queryOp, tags := utils.ExtractQueryTags(rmsg.Message)
	if len(tags) == 0 {
		return nil, nil
	}

	const queryORSelf = "SELECT ctime, message FROM bookmark WHERE userbase_id=$1 AND room_id=$2 AND $3 && (tags || keywords) AND user_id=$5 ORDER BY ctime DESC LIMIT $4"
	const queryANDSelf = "SELECT ctime, message FROM bookmark WHERE userbase_id=$1 AND room_id=$2 AND $3 <@ (tags || keywords) AND user_id=$5 ORDER BY ctime DESC LIMIT $4"

	const queryORRoom = "SELECT ctime, message FROM bookmark WHERE userbase_id=$1 AND room_id=$2 AND $3 && (tags || keywords) ORDER BY ctime DESC LIMIT $4"
	const queryANDRoom = "SELECT ctime, message FROM bookmark WHERE userbase_id=$1 AND room_id=$2 AND $3 <@ (tags || keywords) ORDER BY ctime DESC LIMIT $4"

	var err error
	var rows *sql.Rows
	switch url.Query().Get("scope") {
	case "self", "me", "mine", "myself":
		if queryOp == '+' {
			rows, err = p.db.Query(queryANDSelf, scope[0], scope[1], pq.Array(tags), resultLimit+1, rmsg.Sender.ID)
		} else {
			rows, err = p.db.Query(queryORSelf, scope[0], scope[1], pq.Array(tags), resultLimit+1, rmsg.Sender.ID)
		}
	case "room", "chatroom", "conversation":
		fallthrough
	default:
		if queryOp == '+' {
			rows, err = p.db.Query(queryANDRoom, scope[0], scope[1], pq.Array(tags), resultLimit+1)
		} else {
			rows, err = p.db.Query(queryORRoom, scope[0], scope[1], pq.Array(tags), resultLimit+1)
		}
	}

	// Query failure, nothing much to do!
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
		var resultHeader string

		// Provide hints if results are truncated due to search limit
		if resultCount > resultLimit {
			resultHeader = fmt.Sprintf("showing %d results, try refining search:\n", resultLimit)
		} else {
			resultHeader = fmt.Sprintf("showing %d results:\n", resultCount)
		}

		smsg := types.SendMsg{resultHeader + buff.String(), "text"}
		err = relay.RelayMsg(target, &smsg)
	}

	return nil, err
}

func (p *searchPlugin) Refresh() error {
	return nil
}
