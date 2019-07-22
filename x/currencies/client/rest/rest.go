package rest

import (
    "github.com/cosmos/cosmos-sdk/client/context"
    "github.com/cosmos/cosmos-sdk/codec"
    "github.com/gorilla/mux"
    "fmt"
    "net/http"
    "github.com/cosmos/cosmos-sdk/types/rest"
    "wings-blockchain/x/currencies/types"
    sdk "github.com/cosmos/cosmos-sdk/types"
    "wings-blockchain/x/currencies/msgs"
    "strconv"
    msMsgs "wings-blockchain/x/multisig/msgs"
    sdkRest "github.com/cosmos/cosmos-sdk/client/rest"
)

type msIssueCurrencyReq struct {
    BaseReq   rest.BaseReq `json:"base_req"`
    Symbol    string       `json:"symbol"`
    Amount    string       `json:"amount"`
    Decimals  string       `json:"decimals"`
    Recipient string       `json:"recipient"`
    IssueID   string       `json:"issueID"`
    UniqueID  string       `json:"uniqueID"`
}

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
    r.HandleFunc(fmt.Sprintf("/%s/issue/{issueID}", types.ModuleName), getIssue(cdc, cliCtx)).Methods("GET")
    r.HandleFunc(fmt.Sprintf("/%s/get/{symbol}", types.ModuleName), getCurrency(cdc, cliCtx)).Methods("GET")
    r.HandleFunc(fmt.Sprintf("/%s/issue", types.ModuleName), issueCurrency(cdc, cliCtx)).Methods("POST")
}

func getIssue(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
        issueID := vars["issueID"]

        res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/issue/%s", types.ModuleName, issueID), nil)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        rest.PostProcessResponse(w, cdc, res, true)
    }
}

func getCurrency(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vars   := mux.Vars(r)
        symbol := vars["symbol"]

        res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/get/%s", types.ModuleName, symbol), nil)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        rest.PostProcessResponse(w, cdc, res, true)
    }
}

func issueCurrency(cdc *codec.Codec, cliCtx context.CLIContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var req msIssueCurrencyReq

        if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
            rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
            return
        }

        baseReq := req.BaseReq.Sanitize()

        if !baseReq.ValidateBasic(w) {
            return
        }

        recipient, err := sdk.AccAddressFromBech32(req.Recipient)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        validator, err := sdk.AccAddressFromBech32(baseReq.From)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        d, err := strconv.Atoi(req.Decimals)

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        decimals := int8(d)

        amount, isOk := sdk.NewIntFromString(req.Amount)

        if !isOk {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, "cant parse amount")
            return
        }

        msg := msgs.NewMsgIssueCurrency(
            req.Symbol,
            amount,
            decimals,
            recipient,
            req.IssueID,
        )

        call := msMsgs.NewMsgSubmitCall(msg, req.UniqueID, validator)

        err = call.ValidateBasic()

        if err != nil {
            rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
            return
        }

        sdkRest.WriteGenerateStdTxResponse(w, cdc, cliCtx, baseReq, []sdk.Msg{call})
    }
}
