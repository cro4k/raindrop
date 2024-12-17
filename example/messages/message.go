package messages

import "encoding/json"

type Message struct {
	To      string `json:"to"`
	From    string `json:"from"`
	Message string `json:"message"`
}

func (m Message) JSON() []byte {
	b, _ := json.Marshal(m)
	return b
}
