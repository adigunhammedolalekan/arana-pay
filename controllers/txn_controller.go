package controllers

import (
	"github.com/gin-gonic/gin"
	u "litepay/util"
	"litepay/models"
	"github.com/rpip/paystack-go"
	"os"
)

var InitWalletTopUp = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	id, ok := user . (uint)
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	request := models.WalletTopUpRequest{}
	err := c.ShouldBind(&request)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	account := models.GetAccount(id)
	if account == nil {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	txn := &paystack.TransactionRequest{}
	txn.Amount = request.AmountValueInKobo()
	txn.Email = account.Email

	ps := paystack.NewClient(os.Getenv("PS_KEY"), nil)
	response, err := ps.Transaction.Initialize(txn)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	accessCode, ok := response["access_code"] . (string)
	if !ok {
		c.AbortWithStatusJSON(200, u.Message(false, "Failed to top up wallet at this time. Please retry"))
		return
	}

	r := u.Message(true, "success")
	r["data"] = accessCode
	c.JSON(200, r)
}

var VerifyTransaction = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}
	acc, ok := user . (uint)
	if !ok {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	account := models.GetAccount(acc)
	if account == nil {
		c.AbortWithStatusJSON(403, u.UnAuthorizedMessage())
		return
	}

	ref := c.Param("ref")
	if ref == "" {
		c.JSON(200, u.Message(false, "Invalid transaction reference"))
		return
	}

	txn, err := models.VerifyTransaction(ref)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, "Failed to verify transaction at this time. Please, retry"))
		return
	}

	txRef := models.GetTxRef(ref)
	if txRef != nil {
		c.AbortWithStatusJSON(400,u.Message(false, "Attempt to reuse an" +
			" already used transaction reference"))
		return
	}

	if txn.Status == "success" {
		amountInNaira := txn.Amount / 100
		models.FundAccount(ref, account, amountInNaira)

		if txn.Authorization.Resusable {

			auth := &models.AuthorizationCode{}
			auth.Code = txn.Authorization.AuthorizationCode
			auth.Email = account.Email

			_ = models.CreateAuthorization(auth)
		}
	}

	var message string = "Failed to verify transaction. Please retry"
	if txn.Status == "success" {
		message = "Transaction verification successful"
	}

	c.JSON(200, u.Message(txn.Status == "success", message))
}


var InitPay = func(c *gin.Context) {

	user, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}
	id, ok := user . (uint)
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	token, err := models.CreateToken(id)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	response := u.Message(true, "success")
	response["data"] = token
	c.JSON(200, response)
}

var Pay = func(c *gin.Context) {

	id, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	user, ok := id . (uint)
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	data := &models.RequestPaymentPayload{}
	err := c.ShouldBind(&data)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	err = models.RedeemToken(user, data.Token, data.AmountValue())
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	c.JSON(200, u.Message(true, "success"))
}

var AuthorizePayment = func(c *gin.Context) {

	payload := &models.AuthorizePaymentPayload{}
	err := c.ShouldBind(payload)
	if err != nil {
		c.AbortWithStatusJSON(200, u.InvalidRequestMessage())
		return
	}

	id, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	user, ok := id . (uint)
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	err = models.AuthorizePayment(user, payload)
	if err != nil {
		c.AbortWithStatusJSON(200, u.Message(false, err.Error()))
		return
	}

	c.JSON(200, u.Message(true, "success"))
}

var TxnHistory = func(c *gin.Context) {

	id, ok := c.Get("user")
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	user, ok := id . (uint)
	if !ok {
		c.AbortWithStatusJSON(200, u.UnAuthorizedMessage())
		return
	}

	data := models.GetTransactionHistory(user)
	r := u.Message(true, "success")
	r["data"] = data
	c.JSON(200, data)
}
