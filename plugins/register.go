package plugins

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"bitbucket.org/psyche/types"
)

type registerPlugin struct {
	db      types.DBH
	plugins Psyches
}

type registerMsg struct {
	UserbaseId string
	RoomId     string
	Key        string
	URL        string
	Name       string
}

// Sanitize the input to extract key-value pairs
var sanitizeInputRx = regexp.MustCompile("[ \t]*=[ \t]*")

func NewRegisterPlugin(db *sql.DB, p Psyches) Psyche {
	r := &registerPlugin{types.DBH{db}, p}

	// FIXME: DB admin job in the absence of shell access, devise a better approach for one-off jobs
	// r.db.Exec("DROP TABLE rooms")

	_, err := r.db.Exec("CREATE TABLE IF NOT EXISTS rooms (userbase_id text, room_id text, room_key text, room_url text, room_name text, tags text[], PRIMARY KEY (userbase_id, room_id))")
	if err != nil {
		return nil
	}

	// Register error stream
	rmsg := types.RecvMsg{}
	rmsg.Message = "url=https://botnana.domain.dev.atlassian.io/message?secret=9522becdc4600be22dcf7f6ba12bcf8b657b09f6308478db7056bcaf4c303e688c831d5e3cad8424 name=psyche_error_stream"
	rmsg.Context = "error:error"
	rmsg.Sender.ID = "error"

	u := url.URL{}
	u.Query().Add("room", "true")
	r.Handle(&u, &rmsg)

	return r
}

func (p *registerPlugin) Handle(u *url.URL, rmsg *types.RecvMsg) (smsg *types.SendMsg, err error) {
	// Context: userbaseID:chatroomID/AAID
	scope := strings.SplitN(rmsg.Context, ":", 2)
	if len(scope) != 2 {
		return nil, types.ErrBookmark{fmt.Errorf("missing userbase:chatroom/aaid for scope")}
	}

	// Extract key=value pairs from the message
	rmsg.Message = sanitizeInputRx.ReplaceAllString(rmsg.Message, "=")
	fields := strings.Fields(rmsg.Message)
	var options = make(map[string]string)
	for _, f := range fields {
		// There can be embedded '=' in the value and we do not want to split them
		kv := strings.SplitN(f, "=", 2)
		if len(kv) != 2 {
			continue
		}

		// Normalize the key to lower case
		options[strings.ToLower(kv[0])] = kv[1]
	}

	var msg registerMsg
	msg.UserbaseId = scope[0]
	msg.RoomId = scope[1]

	// The POST URL for the room without which there is nothing much to do
	if v, ok := options["url"]; ok {
		msg.URL = v
	} else {
		return nil, types.ErrRegister{Err: fmt.Errorf("missing key/url in %s", rmsg.Message)}
	}

	// Used for relaying messages, use this as target in supporting endpoints
	if v, ok := options["key"]; ok {
		msg.Key = v
	} else if len(u.Query().Get("room")) > 0 {
		msg.Key = rmsg.Context
	} else {
		msg.Key = msg.UserbaseId + ":" + rmsg.Sender.ID
	}

	if v, ok := options["name"]; ok {
		msg.Name = v
	}

	// Validate the URL and other inputs
	if err := validateURL(msg); err != nil {
		return nil, err
	}

	// Trigger a refresh in affected plugins
	if rp, ok := p.plugins["relay"]; ok {
		defer rp.Refresh()
	}

	// Update if entry exists
	var res sql.Result

	if len(msg.Name) == 0 {
		res, err = p.db.Exec("UPDATE rooms SET room_key=$3, room_url=$4 WHERE userbase_id=$1 AND room_id=$2",
			msg.UserbaseId, msg.RoomId, msg.Key, msg.URL)
	} else {
		res, err = p.db.Exec("UPDATE rooms SET room_key=$3, room_url=$4, room_name=$5 WHERE userbase_id=$1 AND room_id=$2",
			msg.UserbaseId, msg.RoomId, msg.Key, msg.URL, msg.Name)
	}

	if err != nil {
		return nil, err
	}

	if count, err := res.RowsAffected(); err != nil || count > 0 {
		return nil, err
	}

	// Insert if entry does not exist
	_, err = p.db.Exec("INSERT INTO rooms (userbase_id, room_id, room_key, room_url, room_name) SELECT $1, $2, $3, $4, $5 WHERE NOT EXISTS (SELECT 1 FROM rooms WHERE userbase_id=$1 AND room_id=$2)",
		msg.UserbaseId, msg.RoomId, msg.Key, msg.URL, msg.Name)

	return nil, err
}

func (p *registerPlugin) Refresh() error {
	return nil
}

func validateURL(msg registerMsg) (err error) {
	// Post the response to registered room URL
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(types.NewSendMsg(fmt.Sprintf("Psyche room registration invoked with key %s", msg.Key)))
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
