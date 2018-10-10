package controllers

import (
	"github.com/gin-gonic/gin"
	"litepay/models"
	u "litepay/util"
)

var NewAccount = func(c *gin.Context) {

	account := &models.Account{}
	err := c.ShouldBind(account)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	acc, err := models.CreateAccount(account.Email, account.Fullname, account.Password)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	response := u.Message(true, "success")
	response["data"] = acc
	c.JSON(200, response)
}

var Authenticate = func(c *gin.Context) {

	account := &models.Account{}
	err := c.ShouldBind(account)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	acc, err := models.AuthenticateUser(account.Email, account.Password)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	response := u.Message(true, "success")
	response["data"] = acc
	c.JSON(200, response)
}

var CreatePin = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}
	acc := user . (uint)
	account := models.GetAccount(acc)
	if account == nil {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	pin := data["pin"] . (string)
	err = models.CreatePin(account.ID, pin)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, "Failed to create pin at this time"))
		return
	}

	c.JSON(200,u.Message(true, "Pin created"))
}

var VerifyPin = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	acc := user . (uint)
	account := models.GetAccount(acc)
	if account == nil {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	data := make(map[string] interface{})
	err := c.ShouldBind(&data)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	pin := data["pin"] . (string)
	err = models.VerifyPin(account.ID, pin)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	c.JSON(200, u.Message(true, "success"))
}

var GetWallet = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	account, ok := user . (uint)
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	wallet := models.GetWallet(account)
	r := u.Message(true, "success")
	r["data"] = wallet
	c.JSON(200, r)
}

var AddCard = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	account, ok := user . (uint)
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	card := &models.Card{}
	err := c.ShouldBind(card)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	card.Account = account
	err = models.AddCard(card)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	c.JSON(200, u.Message(true, "success"))
}

var GetCards = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	account, ok := user . (uint)
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	data := models.GetCardsFor(account)
	r := u.Message(true, "success")
	r["data"] = data
	c.JSON(200, r)
}

