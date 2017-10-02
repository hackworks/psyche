package plugins

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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
	// Extract key=value pairs from the message
	fields := strings.Fields(rmsg.Message)
	var options = make(map[string]string)
	for _, f := range fields {
		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			continue
		}

		options[strings.ToLower(kv[0])] = kv[1]
	}

	var msg registerMsg

	// The POST URL for the room without which there is nothing much to do
	if v, ok := options["url"]; ok {
		msg.URL = v
	} else {
		return nil, types.ErrRegister{Err: fmt.Errorf("missing key/url in %s", rmsg.Message)}
	}

	// Used for relaying messages, use this as target in supporting endpoints
	if v, ok := options["key"]; ok {
		msg.Key = v
	} else {
		msg.Key = rmsg.Sender.ID
	}

	// A description for the room - helps with human readable output in relayed messages
	if v, ok := options["name"]; ok {
		msg.Name = v
	} else if v, ok := options["description"]; ok {
		msg.Name = v
	} else {
		msg.Name = "Unnamed room"
	}

	// Validate the URL and other inputs
	if err := validateURL(msg); err != nil {
		return nil, err
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

func validateURL(msg registerMsg) (err error) {
	str := fmt.Sprintf("Psyche room registration invoked by %s and url %s", msg.Key, msg.URL)

	// Post the response to registered room URL
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(types.NewSendMsg(str))
	if err != nil {
		return types.ErrRelay{Err: fmt.Errorf("failed to encode response body with error %s", err)}
	}

	resp, err := http.Post(msg.URL, "application/json", body)
	if err == nil {
		defer resp.Body.Close()
	}

	if err != nil {
		err = types.ErrRelay{Err: fmt.Errorf("http post to %s failed with error %s", msg.URL, err)}
	} else if resp.StatusCode != http.StatusOK {
		err = types.ErrRelay{Err: fmt.Errorf("http post to %s returned error %s", msg.URL, resp.Status)}
	}

	return err
}
