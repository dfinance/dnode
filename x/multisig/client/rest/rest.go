package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

const (
	CallID   = "callID"
	UniqueID = "uniqueID"
)

// RegisterRoutes adds endpoint to REST router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/call/{%s}", types.ModuleName, CallID), getCall(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/calls", types.ModuleName), getCalls(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/unique/{%s}", types.ModuleName, UniqueID), getCallByUnique(cliCtx)).Methods("GET")
}

// GetCalls godoc
// @Tags Multisig
// @Summary Get active calls
// @Description Get active call objects
// @ID multisigGetCalls
// @Accept  json
// @Produce json
// @Success 200 {object} MSRespGetCalls
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /multisig/calls [get]
func getCalls(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryCalls), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetCall godoc
// @Tags Multisig
// @Summary Get call
// @Description Get call object by it's ID
// @ID multisigGetCall
// @Accept  json
// @Produce json
// @Param callID path uint true "callID"
// @Success 200 {object} MSRespGetCall
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /multisig/call/{callID} [get]
func getCall(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		vars := mux.Vars(r)
		id, err := helpers.ParseDnIDParam(CallID, vars[CallID], helpers.ParamTypeRestPath)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		req := types.CallReq{CallID: id}
		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryCall), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetUniqueCall godoc
// @Tags Multisig
// @Summary Get call
// @Description Get call object by it's uniqueID
// @ID multisigGetUniqueCall
// @Accept  json
// @Produce json
// @Param uniqueID path string true "call uniqueID"
// @Success 200 {object} MSRespGetCall
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /multisig/unique/{uniqueID} [get]
func getCallByUnique(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		vars := mux.Vars(r)

		req := types.CallByUniqueIdReq{UniqueID: vars[UniqueID]}
		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// send request and process response
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", types.ModuleName, types.QueryCallByUnique), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
