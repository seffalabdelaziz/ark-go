package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kristjank/ark-go/cmd/model"
	log "github.com/sirupsen/logrus"
)

type PostDataResponse struct {
	Success  bool                  `json:"success,omitempty"`
	Payments []model.PaymentRecord `json:"payments,omitempty"`
	Error    string                `json:"error,omitempty"`
	Count    int                   `json:"count,omitempty"`
}

func OnlyLocalCallAllowed() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ClientIP() == "127.0.0.1" || c.ClientIP() == "::1" {
			c.Next()
		} else {
			log.Info("Outside call to service mode is not allowed")
			c.AbortWithStatus(http.StatusBadRequest)
		}
	}
}

//GetServerInformation Returns a server statistics
func GetServerInformation(c *gin.Context) {
	c.JSON(200, gin.H{"version": ArkGoStatsServerVersion})
}

//ReceivePaymetLog from blockchain
func ReceivePaymetLog(c *gin.Context) {
	var recv model.PaymentRecord
	err := c.BindJSON(&recv)

	if err != nil {
		log.Error(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}

	if recv.Delegate == "" || recv.DelegatePubKey == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"success": false})
		return
	}

	recv.SourceIP = c.ClientIP()
	err = ArkStatsDB.Save(&recv)
	log.Info("Received and saved paymentrecord log")
	c.JSON(200, gin.H{"success": true, "logID": recv.Pk})

}

//SendPaymentLog Returns a list of peers to client call. Response is in JSON
func SendPaymentLog(c *gin.Context) {
	offset, err := strconv.Atoi(c.DefaultQuery("offset", "0"))

	payments, err := getPayments(offset)

	if err == nil {
		var response PostDataResponse
		response.Success = true
		response.Payments = payments
		response.Count = len(payments)
		c.JSON(200, response)
	} else {
		c.JSON(500, gin.H{"success": false, "message": err.Error()})
	}
}
