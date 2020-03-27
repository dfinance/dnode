package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/dfinance/dnode/x/poa/types"
)

// Registering routes for REST API.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/validators", types.ModuleName), getValidators(cliCtx)).Methods("GET")
}

// GetValidators godoc
// @Tags poa
// @Summary Get validators
// @Description Get validator objects
// @ID poaValidators
// @Accept  json
// @Produce json
// @Success 200 {object} PoaRespGetValidators
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /poa/validators [get]
func getValidators(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/validators", types.ModuleName), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
