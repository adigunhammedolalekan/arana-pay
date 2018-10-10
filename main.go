package main

import (
	"github.com/gin-gonic/gin"
	"os"
	"github.com/olahol/melody"
	"litepay/controllers"
	"litepay/app"
	"litepay/models"
	"encoding/json"
)

func main() {

	r := gin.New()
	m := melody.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(app.GinJwt)

	//gin.SetMode(gin.ReleaseMode)

	r.GET("/ws/connect", func(context *gin.Context) {
		m.HandleRequest(context.Writer, context.Request)
	})

	m.HandleMessage(func(session *melody.Session, bytes []byte) {

		message := &models.IncomingMessage{}
		err := json.Unmarshal(bytes, message)
		if err == nil {
			switch message.Action {
			case "sub":
				models.CreateWsSubscription(message, session)
				break
			}
		}
	})

	g := r.Group("/api")
	g.POST("/user/new", controllers.NewAccount)
	g.POST("/user/login", controllers.Authenticate)
	g.POST("/txn/init", controllers.InitWalletTopUp)
	g.GET("/txn/verify/:ref", controllers.VerifyTransaction)
	g.POST("/me/pin/new", controllers.CreatePin)
	g.POST("/me/pin/verify", controllers.VerifyPin)
	g.GET("/me/payment/init", controllers.InitPay)
	g.POST("/payment/recv", controllers.Pay)
	g.POST("/payment/authorize", controllers.AuthorizePayment).Use(app.RateLimiterMiddleWare())
	g.GET("/me/txn/history", controllers.TxnHistory)
	g.GET("/me/wallet", controllers.GetWallet)
	g.POST("/card/new", controllers.AddCard)
	g.GET("/me/cards", controllers.GetCards)

	port := os.Getenv("PORT")
	if port == "" {
		port = "2307"
	}

	r.Run(":" + port)
}
