package plugins

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"bitbucket.org/psyche/types"
)

type registerPlugin struct {
	db      types.DBH
	plugins Psyches
}

type registerMsg struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

func NewRegisterPlugin(db *sql.DB, p Psyches) Psyche {
	r := &registerPlugin{types.DBH{db}, p}

	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS rooms (room_key text, room_url text, room_name text, PRIMARY KEY (room_key))")
	if err != nil {
		return nil
	}

	return r
}

func (p *registerPlugin) Handle(url *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	reader := strings.NewReader(rmsg.Context)

	var msg registerMsg
	if err := json.NewDecoder(reader).Decode(&msg); err != nil {
		return nil, err
	}

	if len(msg.Key) == 0 || len(msg.URL) == 0 || len(msg.Name) == 0 {
		return nil, types.ErrRegister{fmt.Errorf("missing key/url/name in %s", rmsg.Context)}
	}

	// Trigger a refresh in affected plugins
	defer p.plugins["relay"].Refresh()

	// Update if entry exists
	res, err := p.db.Exec("UPDATE rooms SET room_key=$1, room_url=$2, room_name=$3 WHERE room_key=$1", msg.Key, msg.URL, msg.Name)
	if err != nil {
		return nil, err
	}

	if count, err := res.RowsAffected(); err != nil || count > 0 {
		return nil, err
	}

	// Insert if entry does not exist
	_, err = p.db.Exec("INSERT INTO rooms (room_key, room_url, room_name) SELECT $1, $2, $3 WHERE NOT EXISTS (SELECT 1 FROM rooms WHERE room_key=$1)", msg.Key, msg.URL, msg.Name)

	return nil, err
}

func (p *registerPlugin) Refresh() error {
	return nil
}
