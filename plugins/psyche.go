package plugins

import (
	"net/url"

	"bitbucket.org/psyche/types"
)

type Psyches map[string]Psyche

type Psyche interface {
	Handle(*url.URL, *types.RecvMsg) (*types.SendMsg, error)
	Refresh() error
}
