package messages

import "encoding/json"

type Message struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Content string `json:"content"`
}

func (m *Message) JSON() []byte {
	b, _ := json.Marshal(m)
	return b
}
