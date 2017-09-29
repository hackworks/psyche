package plugins

import (
	"bytes"
	"database/sql"
	"errors"
	"net/url"
	"strings"

	"bitbucket.org/psyche/types"
	"github.com/jdkato/prose/tokenize"
	"github.com/lib/pq"
)

type searchPlugin struct {
	db      types.DBH
	plugins Psyches
}

// NewBookmarkPlugin creates an instance of bookmark plugin implementing Psyche interface
func NewSearchPlugin(db *sql.DB, p Psyches) Psyche {
	r := &searchPlugin{types.DBH{db}, p}

	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS bookmarks (user_id text, room_id text, tags text[], ctime date, message text)")
	if err != nil {
		return nil
	}

	return r
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

	rows, err := p.db.Query("SELECT ctime, message FROM bookmarks WHERE room_id=$1 AND $2 && tags ORDER BY ctime DESC LIMIT 10",
		rmsg.Context, pq.Array(tags))
	if err != nil {
		return nil, err
	}

	var msg, ct string
	var buff bytes.Buffer
	for rows.Next() {
		err = rows.Scan(&ct, &msg)
		if err != nil {
			break
		}

		buff.WriteString(ct)
		buff.Write([]byte(" >\r\n"))
		buff.WriteString(msg)
	}

	smsg := types.SendMsg{buff.String(), "text"}
	relay.RelayMsg(target, &smsg)

	return &smsg, err
}

func (p *searchPlugin) Refresh() error {
	return nil
}
