package models

import (
	"os"
	"github.com/sendgrid/sendgrid-go"
	"github.com/joho/godotenv"
	"time"
	"net/url"
	"net/http"
	"github.com/jinzhu/gorm"
	"fmt"
	"math/rand"
	"encoding/json"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/rpip/paystack-go"
)

var (
	Db *gorm.DB
	SmsQueue = make(chan *SmsRequest, 10)
	MailQueue = make(chan *MailRequest, 10)
)

func init() {

	e := godotenv.Load()
	if e != nil {
		fmt.Print(e)
	}

	dbUri := os.Getenv("DATABASE_URL")
	conn, err := gorm.Open("postgres", dbUri)
	if err != nil {
		fmt.Print(err)
	}

	rand.Seed(time.Now().UnixNano())

	Db = conn
	Db.Debug().AutoMigrate(&Account{}, &TxToken{},
	&Wallet{}, &Pin{}, &Card{}, &TxRef{})

	go MessageWorker()
}

type Token struct {
	UserId uint
	jwt.StandardClaims
}

func MessageWorker() {

	for {
		select {
		case m, ok := <- SmsQueue:
			if ok && m != nil {
				go m.Send()
			}
			break
		case req, ok := <- MailQueue:
			if ok && req != nil {
				go req.Send()
			}
			break
		}
	}
}

func GetDB() *gorm.DB {
	return Db
}

type AuthCode struct {
	Code json.Number `json:"code"`
}

type MailRequest struct {

	Subject string `json:"subject"`
	Body string `json:"body"`
	To string `json:"to"`
	Name string `json:"name"`

}

type SmsRequest struct {

	ApiToken string `json:"api_token"`
	To string `json:"to"`
	From string `json:"from"`
	Body string `json:"body"`
	DND string `json:"dnd"`
}

func (smsRequest *SmsRequest) Send() (error) {


	apiUrl := "https://www.bulksmsnigeria.com"
	resource := "/api/v1/sms/create/"
	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String()

	req, _ := http.NewRequest("POST", urlStr, nil)
	data := req.URL.Query()
	data.Add("api_token", smsRequest.ApiToken)
	data.Add("to", smsRequest.To)
	data.Add("from", smsRequest.From)
	data.Add("body", smsRequest.Body)
	data.Add("dnd", "1")

	req.URL.RawQuery = data.Encode()
	cli := &http.Client{}
	_, err := cli.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	return nil
}

func (request *MailRequest) Send() (error) {
	return SendEmail(request)
}

func SendEmail(request *MailRequest) error {

	from := mail.NewEmail("LitePay", os.Getenv("email"))
	to := mail.NewEmail(request.Name, request.To)

	message := mail.NewSingleEmail(from, request.Subject, to, request.Body, request.Body)
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	_, err := client.Send(message)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

func GenJWT(user uint) string {
	tk := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), &Token{UserId: user})
	token, _ := tk.SignedString([]byte(os.Getenv("tk_password")))
	return token
}


func VerifyTransaction(ref string) ( *paystack.Transaction, error ) {
	ps := paystack.NewClient(os.Getenv("PS_KEY"), nil)
	return ps.Transaction.Verify(ref)
}
