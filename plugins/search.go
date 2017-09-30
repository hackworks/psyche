package plugins

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

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

	tags := tokenize.NewTreebankWordTokenizer().Tokenize(rmsg.Message)

	// If there are no tags, bail out
	if len(tags) == 0 {
		return nil, nil
	}

	rows, err := p.db.Query("SELECT TO_CHAR(ctime, 'MM-DD-YYYY'), message FROM bookmarks WHERE room_id=$1 AND $2 && tags ORDER BY ctime DESC LIMIT 100",
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

		buff.WriteString(fmt.Sprintf("\n%s >\n%s\n", ct, msg))
	}

	if buff.Len() > 0 {
		smsg := types.SendMsg{buff.String(), "text"}
		relay.RelayMsg(target, &smsg)
	}

	return nil, err
}

func (p *searchPlugin) Refresh() error {
	return nil
}
