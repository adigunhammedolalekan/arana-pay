package models

import (
	"litepay/util"
	"fmt"
	"github.com/olahol/melody"
	"encoding/json"
)

var (
	Alphabets = []string {"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P",
							"R", "S", "T", "U", "V", "W", "X", "Z"}

	sessions = make(map[string] *melody.Session)
)

func GenUniqueKey() (string) {

	var result string = ""
	var i int = 0
	for i < 20 {
		result += Alphabets[util.RandInt(24, 1)]
		i++
	}

	return result[17:] + fmt.Sprintf("-%d-%d", util.RandInt(99999, 1000), util.RandInt(99999, 1000))
}

type IncomingMessage struct {
	Action string `json:"action"`
	UniqueId uint `json:"unique_id"`
}

type WsMessage struct {

	Account *Account `json:"account"`
	Amount float64 `json:"amount"`
	Token string `json:"token"`

}

func CreateWsSubscription(ws *IncomingMessage, sess *melody.Session) {
	key := fmt.Sprintf("account%d", ws.UniqueId)
	sessions[key] = sess
}

func SendWsMessageTo(user uint, message *WsMessage) {

	data, _ := json.Marshal(message)
	key := fmt.Sprintf("account%d", user)
	sess, ok := sessions[key]
	if ok && sess != nil {
		err := sess.Write(data)
		fmt.Println(err)
	}
}

type WalletTopUpRequest struct {
	Amount json.Number `json:"amount"`
}

func (w *WalletTopUpRequest) AmountValueInKobo() float32 {

	data, err := w.Amount.Float64()
	if err != nil {
		return 0
	}

	return float32(data * 100)
}



