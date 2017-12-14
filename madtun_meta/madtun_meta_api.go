package madtun_meta

import (
	"encoding/json"
	"net/url"
)

// ---------------------------------------------------------------------------

type MsgClientInfo struct {
}

type MsgClientInfoResp struct {
	Connect URL
}

// ---------------------------------------------------------------------------

type URL struct {
	*url.URL
}

var _ json.Marshaler = &URL{}
var _ json.Unmarshaler = &URL{}

func (u *URL) MarshalJSON() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *URL) UnmarshalJSON(b []byte) error {
	n, err := url.Parse(string(b))
	if err != nil {
		return err
	}
	u.URL = n
	return nil
}
