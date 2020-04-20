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

	"github.com/dfinance/dnode/x/pricefeed/internal/types"
)

const (
	restName        = "assetCode"
	blockHeightName = "blockHeight"
)

type postPriceReq struct {
	BaseReq    rest.BaseReq `json:"base_req" yaml:"base_req"`
	AssetCode  string       `json:"asset_code" example:"dfi"`                                            // Denom
	Price      string       `json:"price" example:"100"`                                                 // BigInt
	ReceivedAt string       `json:"received_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"` // Timestamp Price createdAt
}

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/rawprices", storeName), postPriceHandler(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/rawprices/{%s}/{%s}", storeName, restName, blockHeightName), getRawPricesHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/currentprice/{%s}", storeName, restName), getCurrentPriceHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/assets", storeName), getAssetsHandler(cliCtx, storeName)).Methods("GET")
}

// PostPrice godoc
// @Tags pricefeed
// @Summary Post Asset RawPrice
// @Description Send Asset RawPrice signed Tx
// @ID pricefeedPostPrice
// @Accept  json
// @Produce json
// @Param postRequest body postPriceReq true "PostPrice request with signed transaction"
// @Success 200 {object} PricefeedRespGetAssets
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /pricefeed/rawprices [put]
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
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

// GetRawPrices godoc
// @Tags pricefeed
// @Summary Get RawPrices
// @Description Get RawPrice objects by assetCode and blockHeight
// @ID pricefeedGetRawPrices
// @Accept  json
// @Produce json
// @Param assetCode path string true "asset code"
// @Param blockHeight path int true "block height rawPrices relates to"
// @Success 200 {object} OracleRespGetRawPrices
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /pricefeed/rawprices/{assetCode}/{blockHeight} [get]
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

// GetCurPrice godoc
// @Tags pricefeed
// @Summary Get current Price
// @Description Get current Price by assetCode
// @ID pricefeedGetRawPrices
// @Accept  json
// @Produce json
// @Param assetCode path string true "asset code"
// @Success 200 {object} OracleRespGetPrice
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /pricefeed/currentprice/{assetCode} [get]
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

// GetAssets godoc
// @Tags pricefeed
// @Summary Get Assets
// @Description Get Asset objects
// @ID pricefeedGetAssets
// @Accept  json
// @Produce json
// @Success 200 {object} OracleRespGetAssets
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /pricefeed/assets [get]
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
