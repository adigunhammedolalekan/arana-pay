package models

import (
	"github.com/jinzhu/gorm"
	"strings"
	"github.com/pkg/errors"
	u "litepay/util"
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

type Account struct {
	gorm.Model
	Email string `json:"email"`
	Fullname string `json:"fullname"`
	Phone string `json:"phone"`
	Password string `json:"password"`
	Token string `sql:"-" gorm:"-" json:"token"`
}

func CreateAccount(email, name, password string) (*Account, error) {

	err := u.ValidateFast(email)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Email address %s is invalid", email))
	}

	if len(strings.TrimSpace(name)) < 3 {
		return nil, errors.New("Invalid fullname supplied")
	}

	if len(strings.TrimSpace(password)) < 6 {
		return nil, errors.New("Invalid password. Too weak. Password length should be at least 6 characters")
	}

	temp := &Account{}
	err = Db.Table("accounts").Where("email = ?", email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.New("Failed to create account at this time. Please retry")
	}

	if temp.ID > 0 {
		return nil, errors.New(fmt.Sprintf("Email address '%s' already in use by another user", email))
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	account := &Account{}
	account.Email = email
	account.Fullname = name
	account.Password = string(hashedPassword)

	tx := Db.Begin()
	err = tx.Error
	if err != nil {
		return nil, errors.New("Failed to create account at this time. Please retry")
	}
	err = tx.Create(account).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("Failed to create account at this time. Please retry")
	}

	wallet := NewWallet(account.ID)
	err = tx.Create(wallet).Error
	if err != nil {
		tx.Rollback()
		return nil, errors.New("Failed to create account at this time. Please retry")
	}

	tx.Commit()
	mail := &MailRequest{}
	mail.Body = "Welcome to LitePay. Cashless. Painless. Seamless"
	mail.Subject = "LitePay"
	mail.To = account.Email

	MailQueue <- mail
	account.Token = GenJWT(account.ID)
	account.Password = "" //Erase password

	return account, nil
}

func AuthenticateUser(email, password string) (*Account, error) {

	err := u.ValidateFast(email)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Email address %s is invalid", email))
	}

	account := &Account{}
	err = Db.Table("accounts").Where("email = ?", email).First(account).Error
	if err != nil && err == gorm.ErrRecordNotFound {
		return nil, errors.New(fmt.Sprintf("User with email address '%s' is not yet registered", email))
	}

	if account.ID <= 0 {
		return nil, errors.New(fmt.Sprintf("User with email address '%s' is not yet registered", email))
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return nil, errors.New("Invalid authentication credentials")
	}

	account.Token = GenJWT(account.ID)
	account.Password = ""
	return account, nil
}

func FundAccount(ref string, user *Account, amount float32) (error) {

	if amount <= 0 {
		return errors.New("Amount should be > 0")
	}

	wallet := GetWallet(user.ID)
	if wallet != nil {
		wallet.Balance += float64(amount)

		tx := Db.Begin()
		err := tx.Error
		if err != nil {
			return err
		}

		err = tx.Table("wallets").Where("user_id = ?", user.ID).Update(wallet).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		txRef := &TxRef{}
		txRef.PsTxRef = ref
		txRef.Amount = amount
		txRef.UserId = user.ID

		err = tx.Create(txRef).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		tx.Commit()
		mail := &MailRequest{}
		mail.Body = "Your account has been funded successfully. Amount = " + fmt.Sprintf("%2f", amount)
		mail.Subject = "LitePay - Account Funded"
		mail.To = user.Email

		MailQueue <- mail
	}

	return nil
}

func AuthorizePayment(user uint, payload *AuthorizePaymentPayload) error {

	token := GetTxToken(payload.Token)
	if token == nil || token.Recv == nil {
		return errors.New(fmt.Sprintf("Token %s not found", payload.Token))
	}

	if token.Status == 1 {
		return errors.New("Token has already been redeemed")
	}

	if user != token.UserId { //Someone else tried to authorize a token that doesn't belong to them
		return errors.New("unAuthorized")
	}

	err := VerifyPin(user, payload.Pin)
	if err != nil {
		return err
	}

	userWallet := GetWallet(token.User.ID)
	recvWallet := GetWallet(token.Recv.ID)

	if userWallet.Balance < token.Amount {
		return errors.New("Insufficient funds")
	}

	//locked area
	tx := Db.Begin()
	err = tx.Error
	if err != nil {
		return err
	}

	userWallet.Balance -= token.Amount
	recvWallet.Balance += token.Amount

	err = tx.Table("wallets").Where("id = ?", userWallet.ID).UpdateColumn("balance",
		userWallet.Balance).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Table("wallets").Where("id = ?", recvWallet.ID).UpdateColumn("balance",
		recvWallet.Balance).Error

	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Table("tx_tokens").Where("token = ?", token.Token).UpdateColumn("status", 1).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func GetAccount(user uint) (*Account) {

	account := &Account{}
	err := Db.Table("accounts").Where("id = ?", user).First(account).Error
	if err != nil {
		return nil
	}

	return account
}

type AuthorizationCode struct {
	gorm.Model
	Code string `json:"code"`
	Email string `json:"email"`
}

func CreateAuthorization(auth *AuthorizationCode) error {

	temp := &AuthorizationCode{}
	err := Db.Table("authorization_codes").Where("email = ?", auth.Email).First(temp).Error
	if err == gorm.ErrRecordNotFound {
		return Db.Create(auth).Error
	}

	return Db.Table("authorization_codes").Where("email = ?", auth.Email).UpdateColumn("code", auth.Code).Error
}

type Pin struct {
	gorm.Model
	Pin string `json:"pin"`
	UserId uint `json:"user_id"`
}

func CreatePin(user uint, code string) error {

	exists := PinExists(user)
	if exists {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
		return Db.Table("pins").Where("user_id", user).UpdateColumn("code", string(hashed)).Error
	}

	pin := &Pin{}
	hashed, _ := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	pin.Pin = string(hashed)
	pin.UserId = user

	return Db.Create(pin).Error
}

func PinExists(user uint) bool {

	var count int = 0
	err := Db.Table("pins").Where("user_id = ?", user).Count(&count).Error
	if err != nil {
		return false
	}

	return count > 0
}

func VerifyPin(user uint, code string) (error) {

	pin := &Pin{}
	err := Db.Table("pins").Where("user_id = ?", user).First(pin).Error
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(pin.Pin), []byte(code))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return errors.New("Pin is invalid/incorrect")
	}

	return nil
}

type Card struct {
	gorm.Model
	Account uint `json:"account"`
	CardNo string `json:"card_no"`
	Cvv string `json:"cvv"`
	ExpiryMonth string `json:"expiry_month"`
	ExpiryYear string `json:"expiry_year"`
}

func AddCard(card *Card) error {

	if card.Account <= 0 {
		return errors.New("Account not found")
	}

	if len(card.Cvv) == 0 {
		return errors.New("Invalid cvv")
	}

	return Db.Create(card).Error
}

func GetCardsFor(user uint) []*Card {

	data := make([]*Card, 0)
	err := Db.Table("cards").Where("account = ?", user).Find(&data).Error
	if err != nil {
		return nil
	}

	return data
}