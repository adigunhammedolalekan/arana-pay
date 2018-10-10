package models

import "github.com/jinzhu/gorm"

type Wallet struct {
	gorm.Model
	UserId uint `json:"user_id"`
	Balance float64 `json:"balance"`
}

func NewWallet(user uint) *Wallet {
	
	wallet := &Wallet{}
	wallet.UserId = user
	wallet.Balance = 0

	return wallet
}

func GetWallet(user uint) *Wallet {

	wallet := &Wallet{}
	err := Db.Table("wallets").Where("user_id = ?", user).First(wallet).Error
	if err != nil {
		return nil
	}

	return wallet
}
