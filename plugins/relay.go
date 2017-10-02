package plugins

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
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
	p.roomMapping.Store("garage", &roomInfo{"Dhruva's private room", "https://botnana.domain.dev.atlassian.io/message?secret=86ebf927b1754b8deb759c1c29701f4bb3ed5bb50eecac981fe9bcb26a733700745c46c4550c2ef8"})
	p.roomMapping.Store("perms_dev", &roomInfo{"Perms Dev", "https://botnana.domain.dev.atlassian.io/message?secret=e74648426f6e0c67b750c7ebb7aa021f334bce305b19aeea562f72dbbf02fad59e1370f77b662332"})
	p.roomMapping.Store("permissions_service", &roomInfo{"Permissions Service", "https://botnana.domain.dev.atlassian.io/message?secret=126b4ff64e38d95929924a2b3527b24512ae8e366149d41f168bd0a84b8ddc8dcc07d4cf35ad4c17"})
	p.roomMapping.Store("triforce", &roomInfo{"Triforce (MTV Identity)", "https://botnana.domain.dev.atlassian.io/message?secret=c80c4ee687de9b95784916882d005a9e69f2ce91c76604a68e77cbdda2da79690ab5987058fc1aea"})

	p.Refresh()
}

func (p *relayPlugin) getResponse(source string, rmsg *types.RecvMsg) *types.SendMsg {
	return types.NewSendMsg(fmt.Sprintf("Message from room %s: %s?", source, rmsg.Message))
}

func (p *relayPlugin) Handle(url *url.URL, rmsg *types.RecvMsg) (*types.SendMsg, error) {
	source := url.Query().Get("source")
	target := url.Query().Get("target")

	sourceRoom := "Unknown"
	if val, ok := p.roomMapping.Load(source); ok {
		s, _ := val.(*roomInfo)
		sourceRoom = s.name
	}

	// Get the response to relay
	smsg := p.getResponse(sourceRoom, rmsg)

	return smsg, p.RelayMsg(target, smsg)
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

func (p *relayPlugin) RelayMsg(target string, smsg *types.SendMsg) error {
	val, ok := p.roomMapping.Load(target)
	if !ok {
		return types.ErrRelay{errors.New("target room to send results missing")}
	}

	room, ok := val.(*roomInfo)
	if !ok {
		return types.ErrRelay{fmt.Errorf("target room mapping missing for %s", target)}
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
