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
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

const (
	OrderID        = "orderID"
	OrderOwner     = "owner"
	OrderDirection = "direction"
	OrderMarketID  = "marketID"
)

type PostOrderReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	// Market assetCode in the following format: {base_denomination_symbol}_{quote_denomination_symbol}
	AssetCode dnTypes.AssetCode `json:"asset_code" example:"btc_dfi"`
	// Order type (ask/bid)
	Direction types.Direction `json:"direction" example:"ask"`
	// QuoteAsset price with decimals (1.0 DFI with 18 decimals -> 1000000000000000000)
	Price string `json:"price" example:"100"`
	// BaseAsset quantity with decimals (1.0 BTC with 8 decimals -> 100000000)
	Quantity string `json:"quantity" example:"10"`
	// Order TTL [s]
	TtlInSec string `json:"ttl_in_sec" example:"3"`
}

type RevokeOrderReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	OrderID string       `json:"order_id" yaml:"order_id" example:"100"`
}

// RegisterRoutes adds endpoint to REST router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s", types.ModuleName), getOrdersWithParams(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/{%s}", types.ModuleName, OrderID), getOrder(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/post", types.ModuleName), postOrder(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/revoke", types.ModuleName), revokeOrder(cliCtx)).Methods("PUT")
}

// GetOrdersWithParams godoc
// @Tags Orders
// @Summary Get orders
// @Description Get array of Order objects with pagination and filters
// @ID ordersGetOrdersWithParams
// @Accept  json
// @Produce json
// @Param page query int false "page number (first page: 1)"
// @Param limit query int false "items per page (default: 100)"
// @Param owner query string false "owner filter"
// @Param direction query string false "direction filter (bid/ask)"
// @Param marketID query string false "marketID filter (bid/ask)"
// @Success 200 {object} OrdersRespGetOrders
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query/path params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /orders [get]
func getOrdersWithParams(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		pageStr := r.URL.Query().Get("page")
		limitStr := r.URL.Query().Get("limit")
		page, limit, err := helpers.ParsePaginationParams(pageStr, limitStr, helpers.ParamTypeRestQuery)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		ownerFilterStr := r.URL.Query().Get(OrderOwner)
		directionFilterStr := r.URL.Query().Get(OrderDirection)
		marketIDFitler := r.URL.Query().Get(OrderMarketID)

		ownerFilter := sdk.AccAddress{}
		if ownerFilterStr != "" {
			var err error
			ownerFilter, err = sdk.AccAddressFromBech32(ownerFilterStr)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%s param parsing: %v", OrderOwner, err))
				return
			}
		}

		// prepare request
		req := types.OrdersReq{
			Page:      page,
			Limit:     limit,
			Owner:     ownerFilter,
			Direction: types.NewDirectionRaw(directionFilterStr),
			MarketID:  marketIDFitler,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// query and parse the result
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/list", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// GetOrder godoc
// @Tags Orders
// @Summary Get order
// @Description Get Order object by orderID
// @ID ordersGetOrder
// @Accept  json
// @Produce json
// @Param orderID path string true "orderID"
// @Success 200 {object} OrdersRespGetOrder
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query/path params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /orders/{orderID} [get]
func getOrder(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		vars := mux.Vars(r)
		id, err := helpers.ParseDnIDParam(OrderID, vars[OrderID], helpers.ParamTypeRestPath)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// prepare request
		req := types.OrderReq{
			ID: id,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// query and parse the result
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/order", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}

// postOrder godoc
// @Tags Orders
// @Summary Post new order
// @Description Post new order
// @ID ordersPostOrder
// @Accept  json
// @Produce json
// @Param postRequest body PostOrderReq true "PostOrder request with signed transaction"
// @Success 200 {object} OrdersRespPostOrder
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /orders/post [put]
func postOrder(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		var req PostOrderReq
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

		if err := req.AssetCode.Validate(); err != nil {
			err := helpers.BuildError("asset_code", req.AssetCode.String(), helpers.ParamTypeRestRequest, err.Error())
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if !req.Direction.IsValid() {
			err := helpers.BuildError("direction", req.Direction.String(), helpers.ParamTypeRestRequest, types.ErrWrongDirection.Error())
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		price, err := helpers.ParseSdkUintParam("price", req.Price, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		quantity, err := helpers.ParseSdkUintParam("quantity", req.Quantity, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		ttl, err := helpers.ParseUint64Param("ttl_in_sec", req.TtlInSec, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// prepare and send msg
		msg := types.NewMsgPost(fromAddr, req.AssetCode, req.Direction, price, quantity, ttl)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

// revokeOrder godoc
// @Tags Orders
// @Summary Revoke order
// @Description Revoke order
// @ID ordersRevokeOrder
// @Accept  json
// @Produce json
// @Param postRequest body RevokeOrderReq true "RevokeOrder request with signed transaction"
// @Success 200 {object} OrdersRespRevokeOrder
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /orders/revoke [put]
func revokeOrder(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		var req RevokeOrderReq
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

		id, err := helpers.ParseDnIDParam("order_id", req.OrderID, helpers.ParamTypeRestRequest)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		// prepare and send msg
		msg := types.NewMsgRevokeOrder(fromAddr, id)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
