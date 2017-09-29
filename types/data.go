package types

import (
	"database/sql"
	"encoding/json"
	"io"
)

// SendMsg models the message sent to botler via POST
type SendMsg struct {
	Text   string `json:"text"`
	Format string `json:"format"`
}

// RecvMsg models the message received from botler
type RecvMsg struct {
	Message string `json:"message"`
	Context string `json:"context"`
	Sender  struct {
		ID string `json:"id"`
	} `json:"sender"`
}

type DBH struct {
	*sql.DB
}

// NewRecvMsg constructs a RecvMsg from HTTP POST request
func NewRecvMsg(req io.Reader) (*RecvMsg, error) {
	var r RecvMsg
	err := json.NewDecoder(req).Decode(&r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}
