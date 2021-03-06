package web

import (
	"errors"
	"fmt"
	"net/http"

	"chainlink/core/assets"
	"chainlink/core/services/chainlink"
	"chainlink/core/store/models"
	"chainlink/core/store/presenters"
	"chainlink/core/utils"

	"github.com/gin-gonic/gin"
)

// WithdrawalsController can send LINK tokens to another address
type WithdrawalsController struct {
	App chainlink.Application
}

var naz = assets.NewLink(1)

// Create sends LINK from the configured oracle contract to the given address
// See models.WithdrawalRequest for expected inputs. ContractAddress field is
// optional.
//
// Example: "<application>/withdrawals"
func (abc *WithdrawalsController) Create(c *gin.Context) {
	wr := models.WithdrawalRequest{}
	if err := c.ShouldBindJSON(&wr); err != nil {
		jsonAPIError(c, http.StatusBadRequest, err)
		return
	}

	if wr.Amount.Cmp(naz) < 0 {
		err := fmt.Errorf("Must withdraw at least %v LINK", naz.String())
		jsonAPIError(c, http.StatusBadRequest, err)
		return
	}

	addressWasNotSpecifiedInRequest := utils.ZeroAddress
	if wr.DestinationAddress == addressWasNotSpecifiedInRequest {
		jsonAPIError(c, http.StatusBadRequest, errors.New("Invalid withdrawal address"))
		return
	}

	store := abc.App.GetStore()
	txm := store.TxManager
	linkBalance, err := txm.ContractLINKBalance(wr)
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}

	if linkBalance.Cmp(wr.Amount) < 0 {
		jsonAPIError(c, http.StatusBadRequest, fmt.Errorf(
			"Insufficient link balance. Withdrawal Amount: %v "+
				"Link Balance: %v",
			wr.Amount.String(), linkBalance.String()))
		return
	}

	hash, err := txm.WithdrawLINK(wr)
	if err != nil {
		jsonAPIError(c, http.StatusInternalServerError, err)
		return
	}

	jsonAPIResponse(c, presenters.NewTx(&models.Tx{Hash: hash}), "transaction")
}
