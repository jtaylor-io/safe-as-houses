package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/jtaylor-io/safe-as-houses/db/sqlc"
	"github.com/shopspring/decimal"
)

type createAccountRequest struct {
	Owner    string `json:"owner"    binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR HKD"`
}

func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  decimal.Zero,
	}

	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}
