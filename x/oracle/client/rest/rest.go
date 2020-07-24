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
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

const (
	assetCodeKey   = "assetCode"
	blockHeightKey = "blockHeight"
)

type PostPriceReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	// AssetCode
	AssetCode string `json:"asset_code" example:"btc_dfi"`
	// Price in big.Int format
	Price string `json:"price" example:"100"`
	// Timestamp price createdAt
	ReceivedAt string `json:"received_at" format:"RFC 3339" example:"2020-03-27T13:45:15.293426Z"`
}

// RegisterRoutes Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(fmt.Sprintf("/%s/rawprices", storeName), postPriceHandler(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/rawprices/{%s}/{%s}", storeName, assetCodeKey, blockHeightKey), getRawPricesHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/currentprice/{%s}", storeName, assetCodeKey), getCurrentPriceHandler(cliCtx, storeName)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/assets", storeName), getAssetsHandler(cliCtx, storeName)).Methods("GET")
}

// PostPrice godoc
// @Tags Oracle
// @Summary Post asset rawPrice
// @Description Send asset rawPrice signed Tx
// @ID oraclePostPrice
// @Accept  json
// @Produce json
// @Param postRequest body PostPriceReq true "PostPrice request with signed transaction"
// @Success 200 {object} OracleRespGetAssets
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /oracle/rawprices [put]
func postPriceHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		var req PostPriceReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "failed to parse request")
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		addr, err := helpers.ParseSdkAddressParam("from", baseReq.From, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		assetCode, err := helpers.ParseAssetCodeParam("assetCode", req.AssetCode, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		price, err := helpers.ParseSdkIntParam("price", req.Price, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		receivedAt, err := helpers.ParseUnixTimestamp("receivedAt", req.ReceivedAt, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// create the message
		msg := types.NewMsgPostPrice(addr, assetCode, price, receivedAt)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

// GetRawPrices godoc
// @Tags Oracle
// @Summary Get rawPrices
// @Description Get rawPrice objects by assetCode and blockHeight
// @ID oracleGetRawPrices
// @Accept  json
// @Produce json
// @Param assetCode path string true "asset code"
// @Param blockHeight path int true "block height rawPrices relates to"
// @Success 200 {object} OracleRespGetRawPrices
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 404 {object} rest.ErrorResponse "Returned if requested data wasn't found"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /oracle/rawprices/{assetCode}/{blockHeight} [get]
func getRawPricesHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		vars := mux.Vars(r)

		assetCode, err := helpers.ParseAssetCodeParam("assetCode", vars[assetCodeKey], helpers.ParamTypeRestPath)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		blockHeight, err := helpers.ParseUint64Param(blockHeightKey, vars[blockHeightKey], helpers.ParamTypeRestPath)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// send request and process response
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s/%d", storeName, types.QueryRawPrices, assetCode, blockHeight), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetCurPrice godoc
// @Tags Oracle
// @Summary Get current Price
// @Description Get current Price by assetCode
// @ID oracleGetCurrentPrice
// @Accept  json
// @Produce json
// @Param assetCode path string true "asset code"
// @Success 200 {object} OracleRespGetPrice
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 404 {object} rest.ErrorResponse "Returned if requested data wasn't found"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /oracle/currentprice/{assetCode} [get]
func getCurrentPriceHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		vars := mux.Vars(r)

		assetCode, err := helpers.ParseAssetCodeParam("assetCode", vars[assetCodeKey], helpers.ParamTypeRestPath)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// parse inputs and prepare request
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s/%s", storeName, types.QueryPrice, assetCode), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetAssets godoc
// @Tags Oracle
// @Summary Get assets
// @Description Get asset objects
// @ID oracleGetAssets
// @Accept  json
// @Produce json
// @Success 200 {object} OracleRespGetAssets
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 404 {object} rest.ErrorResponse "Returned if requested data wasn't found"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /oracle/assets [get]
func getAssetsHandler(cliCtx context.CLIContext, storeName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs and prepare request
		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// parse inputs and prepare request
		res, height, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/%s", storeName, types.QueryAssets), nil)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusNotFound, err.Error())
			return
		}
		cliCtx = cliCtx.WithHeight(height)

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
