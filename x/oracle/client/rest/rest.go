package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/gorilla/mux"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/WingsDao/wings-blockchain/x/oracle/internal/types"
)

const (
	restName        = "assetCode"
	blockHeightName = "blockHeight"
)

type postPriceReq struct {
	BaseReq    rest.BaseReq `json:"base_req"`
	AssetCode  string       `json:"asset_code"`
	Price      string       `json:"price"`
	ReceivedAt string       `json:"received_at"`
}

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/rawprices", storeName), postPriceHandler(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/rawprices/{%s}/{%s}", storeName, restName, blockHeightName), getRawPricesHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/currentprice/{%s}", storeName, restName), getCurrentPriceHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/assets", storeName), getAssetsHandler(cliCtx, storeName)).Methods("GET")
}

func postPriceHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req postPriceReq

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := sdk.AccAddressFromBech32(baseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		price, isOk := sdk.NewIntFromString(req.Price)
		if !isOk {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("something wrong with price value: %s", req.Price))
			return
		}

		receivedAtInt, ok := sdk.NewIntFromString(req.ReceivedAt)
		if !ok {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "invalid expiry")
			return
		}
		receivedAt := tmtime.Canonical(time.Unix(receivedAtInt.Int64(), 0))

		// create the message
		msg := types.NewMsgPostPrice(addr, req.AssetCode, price, receivedAt)
		err = msg.ValidateBasic()
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

func getRawPricesHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		assetCode := vars[restName]
		blockHeight, err := strconv.ParseInt(vars[blockHeightName], 10, 64)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("invalid blockHeight parameter: %v", err))
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/rawprices/%s/%d", storeName, assetCode, blockHeight), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getCurrentPriceHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		paramType := vars[restName]
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/price/%s", storeName, paramType), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

func getAssetsHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/assets/", storeName), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
