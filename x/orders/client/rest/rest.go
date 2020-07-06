package rest

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders/internal/types"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

const (
	OrderID        = "orderID"
	OrderOwner     = "owner"
	OrderDirection = "direction"
	OrderMarketID  = "marketID"
)

type PostOrderReq struct {
	BaseReq   rest.BaseReq      `json:"base_req" yaml:"base_req"`
	AssetCode dnTypes.AssetCode `json:"asset_code" example:"btc_dfi"`
	Direction types.Direction   `json:"direction" example:"ask"`
	Price     string            `json:"price" example:"100"`
	Quantity  string            `json:"quantity" example:"10"`
	TtlInSec  string            `json:"ttl_in_sec" example:"3"`
}

type RevokeOrderReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
	OrderId string       `json:"order_id" yaml:"order_id"`
}

// RegisterRoutes adds endpoint to REST router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s", types.ModuleName), getOrdersWithParams(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/{%s}", types.ModuleName, OrderID), getOrder(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/post", types.ModuleName), postOrder(cliCtx)).Methods("PUT")
	r.HandleFunc(fmt.Sprintf("/%s/revoke", types.ModuleName), revokeOrder(cliCtx)).Methods("PUT")
}

// GetOrdersWithParams godoc
// @Tags orders
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
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /orders [get]
func getOrdersWithParams(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		page := r.URL.Query().Get("page")
		if page == "" {
			page = "1"
		}
		parsedPage, err := strconv.ParseInt(page, 10, 32)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("%q query param: %v", "page", err))
			return
		}

		limit := r.URL.Query().Get("limit")
		if limit == "" {
			limit = "100"
		}
		parsedLimit, err := strconv.ParseInt(limit, 10, 32)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("%q query param: %v", "limit", err))
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
			Page:      int(parsedPage),
			Limit:     int(parsedLimit),
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
// @Tags orders
// @Summary Get order
// @Description Get Order object by orderID
// @ID ordersGetOrder
// @Accept  json
// @Produce json
// @Param orderID path string true "orderID"
// @Success 200 {object} OrdersRespGetOrder
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /orders/{orderID} [get]
func getOrder(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		vars := mux.Vars(r)
		id, err := dnTypes.NewIDFromString(vars[OrderID])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%q param parsing: %v", OrderID, err))
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
// @Tags orders
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
		var req PostOrderReq

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

		if err := req.AssetCode.Validate(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		if !req.Direction.IsValid() {
			rest.WriteErrorResponse(w, http.StatusBadRequest, types.ErrWrongDirection.Error())
			return
		}

		price, err := sdk.ParseUint(req.Price)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("something wrong with price value: %s", err.Error()))
			return
		}

		quantity, err := sdk.ParseUint(req.Quantity)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("something wrong with quantity value: %s", err.Error()))
			return
		}

		ttl := sdk.NewUintFromString(req.TtlInSec)
		if ttl.IsZero() {
			rest.WriteErrorResponse(w, http.StatusBadRequest, "something wrong with ttl value")
			return
		}

		msg := types.NewMsgPost(addr, req.AssetCode, req.Direction, price, quantity, ttl.Uint64())
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}

// revokeOrder godoc
// @Tags orders
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
		var req RevokeOrderReq

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

		id, err := dnTypes.NewIDFromString(req.OrderId)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%q param parsing: %v", OrderID, err))
			return
		}

		msg := types.NewMsgRevokeOrder(addr, id)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, baseReq, []sdk.Msg{msg})
	}
}
