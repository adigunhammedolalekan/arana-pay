package models

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"fmt"
	"encoding/json"
)

//Represent a transaction token
type TxToken struct {
	gorm.Model
	Token string `json:"token"`
	Amount float64 `json:"amount"`
	Status uint `json:"status"`
	UserId uint `json:"user_id"`
	RecvBy uint `json:"recv_by"`

	User *Account `sql:"-" gorm:"-" json:"user"`
	Recv *Account `sql:"-" gorm:"-" json:"recv"`
}

type RequestPaymentPayload struct {
	Amount json.Number `json:"amount"`
	Token string `json:"token"`
}

type AuthorizePaymentPayload struct {
	Token string `json:"token"`
	Pin string `json:"pin"`
}

func (p *RequestPaymentPayload) AmountValue() float64 {

	data, err := p.Amount.Float64()
	if err != nil {
		return 0
	}

	return data
}

func CreateToken(user uint) (*TxToken, error) {

	account := GetAccount(user)
	if account == nil {
		return nil, errors.New("Account not found")
	}

	token := findToken()
	tx := &TxToken{}
	tx.UserId = user
	tx.Token = token

	err := Db.Create(tx).Error
	if err != nil {
		return nil, errors.New("Cannot create token at this time. Please retry")
	}

	return tx, nil
}

//recursively find a unique token
func findToken() string {

	token := GenUniqueKey()
	tx := &TxToken{}
	err := Db.Table("tx_tokens").Where("token = ?", token).First(tx).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return token
	}

	return findToken()
}

func GetTxToken(token string) *TxToken {

	tx := &TxToken{}
	err := Db.Table("tx_tokens").Where("token = ?", token).First(tx).Error
	if err != nil {
		return nil
	}

	if tx.ID <= 0 {
		return nil
	}

	tx.Recv = GetAccount(tx.RecvBy)
	tx.User = GetAccount(tx.UserId)
	return tx
}

func RedeemToken(user uint, tk string, amount float64) (error) {

	token := GetTxToken(tk)
	if token == nil {
		return errors.New(fmt.Sprintf("Token %s not found", tk))
	}

	wallet := GetWallet(token.UserId)
	if wallet == nil {
		return errors.New("Wallet not found for user")
	}

	if wallet.Balance < amount {
		return errors.New(fmt.Sprintf("The wallet balance on %s account is insufficient to complete this transaction", token.User.Fullname))
	}

	token.Amount = amount
	tx := Db.Begin()
	err := tx.Error
	if err != nil {
		return err
	}

	err = tx.Table("tx_tokens").Where("token = ?", tk).UpdateColumn("amount", amount).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Table("tx_tokens").Where("token = ?", tk).UpdateColumn("recv_by", user).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	wsMessage := &WsMessage{}
	wsMessage.Account = GetAccount(user)
	wsMessage.Amount = amount
	wsMessage.Token = tk

	SendWsMessageTo(token.User.ID, wsMessage)
	return nil
}

func GetTxTokenById(id uint) *TxToken {

	tx := &TxToken{}
	err := Db.Table("tx_tokens").Where("id = ?", id).First(tx).Error
	if err != nil {
		return nil
	}

	if tx.ID <= 0 {
		return nil
	}

	tx.Recv = GetAccount(tx.RecvBy)
	tx.User = GetAccount(tx.UserId)
	return tx
}

func GetTransactionHistory(user uint) []*TxToken {

	data := make([]*TxToken, 0)
	err := Db.Table("tx_tokens").Where("user_id = ? AND status = ? AND amount > ?", user, 1, 0).Or("recv_by = ? AND status = ? AND amount > ?", user, 1, 0).Find(&data).Error
	if err != nil {
		return nil
	}

	resp := make([]*TxToken, 0)
	for _, n := range data {
		resp = append(resp, GetTxToken(n.Token))
	}

	return resp
}

type TxRef struct {
	gorm.Model
	UserId uint `json:"user_id"`
	PsTxRef string `json:"ps_tx_ref"` //PayStack Transaction reference
	Amount float32 `json:"amount"`
}

func CreateTxRef(ref *TxRef) error {
	if len(ref.PsTxRef) == 0 {
		return errors.New("Cannot create TxRef with empty {ref}")
	}

	return Db.Create(ref).Error
}

func GetTxRef(ref string) *TxRef {

	tx := &TxRef{}
	err := Db.Table("tx_refs").Where("ps_tx_ref = ?", ref).First(tx).Error
	if err != nil {
		return nil
	}

	return tx
}