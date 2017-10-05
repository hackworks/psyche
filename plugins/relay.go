package plugins

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"bitbucket.org/psyche/types"
)

type relayPlugin struct {
	db          types.DBH
	roomMapping sync.Map
	plugins     Psyches
}

type roomInfo struct {
	name string
	url  string
}

// NewRelayPlugin returns an instance of message relay Psyche implementation
func NewRelayPlugin(db *sql.DB, p Psyches) Psyche {
	r := &relayPlugin{}

	r.db = types.DBH{db}
	r.plugins = p

	r.init()

	return r
}

func (p *relayPlugin) init() {
	p.Refresh()
}

func (p *relayPlugin) getResponse(source string, rmsg *types.RecvMsg) *types.SendMsg {
	return types.NewSendMsg(fmt.Sprintf("Message from room %s: %s?", source, rmsg.Message))
}

func (p *relayPlugin) Handle(url *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	source := url.Query().Get("source")
	target := url.Query().Get("target")

	if len(source) == 0 {
		source = rmsg.Context
	}

	sourceRoom := "Unnamed room"
	if val, ok := p.roomMapping.Load(source); ok {
		s, _ := val.(*roomInfo)
		sourceRoom = s.name
	}

	// Get the response to relay
	smsg := p.getResponse(sourceRoom, rmsg)

	return smsg, p.RelayMsg(rmsg, target, smsg)
}

func (p *relayPlugin) Refresh() error {
	if p.db.DB == nil {
		return nil
	}

	rows, err := p.db.Query("SELECT room_key, room_url, room_name FROM rooms")
	if err != nil {
		return err
	}

	var room_key, room_url, room_name string
	for rows.Next() {
		if err = rows.Scan(&room_key, &room_url, &room_name); err != nil {
			return err
		}

		p.roomMapping.Store(room_key, &roomInfo{room_name, room_url})
	}

	return rows.Close()
}

func (p *relayPlugin) RelayMsg(rmsg *types.RecvMsg, target string, smsg *types.SendMsg) error {
	// Attempt with given target
	val, ok := p.roomMapping.Load(target)
	if !ok {
		// If we fail, we attempt to send response to room from which we got the request
		target = rmsg.Context
		val, ok = p.roomMapping.Load(target)

		// If we don't have that room registered, we cannot do much, return error and send to "error:error"
		if !ok {
			return types.ErrRelay{fmt.Errorf("target room mapping missing for %s", target)}
		}
	}

	room, ok := val.(*roomInfo)
	if !ok {
		return types.ErrRelay{fmt.Errorf("target room mapping typecasting failed for %s", target)}
	}

	// Prepare response for POST
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(smsg)
	if err != nil {
		return types.ErrRelay{fmt.Errorf("failed to encode response body with error %s", err)}
	}

	// Post the response to registered room URL
	resp, err := http.Post(room.url, "application/json", body)
	if err != nil {
		return types.ErrRelay{fmt.Errorf("http post to %s failed with error %s", room.url, err)}
	}
	resp.Body.Close()

	return nil
}
