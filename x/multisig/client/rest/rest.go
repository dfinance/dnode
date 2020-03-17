// Implements REST API for multisig modules.
package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	"github.com/dfinance/dnode/x/multisig/types"
)

// Registering routes in the REST API.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s/call/{id}", types.ModuleName), getCall(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/calls", types.ModuleName), getCalls(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/unique/{unique}", types.ModuleName), getCallByUnique(cliCtx)).Methods("GET")
}

// Getting call by unique id from REST API.
func getCallByUnique(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		req := types.UniqueReq{UniqueId: vars["unique"]}

		bz, err := cliCtx.Codec.MarshalJSON(req)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/unique", types.ModuleName), bz)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// Get call by id from REST API.
func getCall(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := strconv.ParseUint(vars["id"], 10, 64)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}

		req := types.CallReq{CallId: id}
		bz, err := cliCtx.Codec.MarshalJSON(req)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/call", types.ModuleName), bz)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// Get list of calls from REST API.
func getCalls(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/calls", types.ModuleName), nil)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
