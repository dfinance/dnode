package rest

import (
    "github.com/cosmos/cosmos-sdk/client/context"
    "github.com/gorilla/mux"
    "fmt"
    "wings-blockchain/x/multisig/types"
    "github.com/cosmos/cosmos-sdk/codec"
    "net/http"
    "github.com/cosmos/cosmos-sdk/types/rest"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
    r.HandleFunc(fmt.Sprintf("/%s/call/{id}", types.ModuleName), getCall(cdc, cliCtx)).Methods("GET")
    r.HandleFunc(fmt.Sprintf("/%s/calls", types.ModuleName), getCalls(cdc, cliCtx)).Methods("GET")
    r.HandleFunc(fmt.Sprintf("/%s/unique/{unique}", types.ModuleName), getCallByUnique(cdc, cliCtx)).Methods("GET")

}

func getCallByUnique(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["unique"]

        res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/unique/%s", types.ModuleName, id), nil)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        rest.PostProcessResponse(w, cdc, res, true)
    }
}


func getCall(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        id := vars["id"]

        res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/call/%s", types.ModuleName, id), nil)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        rest.PostProcessResponse(w, cdc, res, true)
    }
}

func getCalls(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        mode := vars["mode"]

        res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/calls/%s", types.ModuleName, mode), nil)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        rest.PostProcessResponse(w, cdc, res, true)
    }
}
