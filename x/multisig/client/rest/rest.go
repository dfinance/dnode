package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

const (
	CallID   = "callID"
	UniqueID = "uniqueID"
)

type ConfirmReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	// Confirming CallID
	CallID string `json:"call_id" yaml:"call_id" example:"0" format:"string representation for big.Uint"`
}

type RevokeReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	// Confirming CallID
	CallID string `json:"call_id" yaml:"call_id" example:"0" format:"string representation for big.Uint"`
}

// RegisterRoutes adds endpoint to REST router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/call/{%s}", types.ModuleName, CallID), getCall(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/calls", types.ModuleName), getCalls(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/unique/{%s}", types.ModuleName, UniqueID), getCallByUnique(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/confirm", types.ModuleName), confirm(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/revoke", types.ModuleName), revoke(cliCtx)).Methods("PUT")
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

// ConfirmCall godoc
// @Tags Multisig
// @Summary Confirm call
// @Description Get confirm multi signature call by PoA validator stdTx object
// @ID multisigConfirm
// @Accept  json
// @Produce json
// @Param request body ConfirmReq true "Confirm request"
// @Success 200 {object} CCRespStdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /multisig/confirm [put]
func confirm(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		var req ConfirmReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := helpers.ParseSdkAddressParam("from", baseReq.From, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		callID, err := helpers.ParseDnIDParam("call_id", req.CallID, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgConfirmCall(callID, fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

// RevokeCall godoc
// @Tags Multisig
// @Summary Revoke call
// @Description Get revoke multi signature call vote by PoA validator stdTx object
// @ID multisigRevoke
// @Accept  json
// @Produce json
// @Param request body RevokeReq true "Revoke request"
// @Success 200 {object} CCRespStdTx
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /multisig/revoke [put]
func revoke(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		var req RevokeReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := helpers.ParseSdkAddressParam("from", baseReq.From, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		callID, err := helpers.ParseDnIDParam("call_id", req.CallID, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgRevokeConfirm(callID, fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
