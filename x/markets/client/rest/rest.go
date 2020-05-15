package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

const (
	MarketID         = "marketID"
	MarketBaseDenom  = "baseAssetDenom"
	MarketQuoteDenom = "quoteAssetDenom"
)

// RegisterRoutes adds endpoint to REST router.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/%s", types.ModuleName), getMarketsWithParams(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/%s/{%s}", types.ModuleName, MarketID), getMarket(cliCtx)).Methods("GET")
}

// GetMarketsWithParams godoc
// @Tags markets
// @Summary Get markets
// @Description Get array of Market objects with pagination and filters
// @ID marketsGetMarketsWithParams
// @Accept  multipart/form-data
// @Produce json
// @Param page formData int false "page number (first page: 1)"
// @Param limit formData int false "items per page (default: 100)"
// @Param baseAssetDenom query string false "BaseAsset denom filter"
// @Param quoteAssetDenom query string false "QuoteAsset denom filter"
// @Success 200 {object} MarketsRespGetMarkets
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /markets [get]
func getMarketsWithParams(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		baseDenomFilter := r.URL.Query().Get(MarketBaseDenom)
		quoteDenomFilter := r.URL.Query().Get(MarketQuoteDenom)

		// prepare request
		req := types.MarketsReq{
			Page:            page,
			Limit:           limit,
			BaseAssetDenom:  baseDenomFilter,
			QuoteAssetDenom: quoteDenomFilter,
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

// GetMarket godoc
// @Tags markets
// @Summary Get market
// @Description Get Market object by marketID
// @ID marketsGetMarket
// @Accept  json
// @Produce json
// @Param marketID path string true "marketID"
// @Success 200 {object} MarketsRespGetMarket
// @Failure 400 {object} rest.ErrorResponse "Returned if the request doesn't have valid query params"
// @Failure 500 {object} rest.ErrorResponse "Returned on server error"
// @Router /markets/{marketID} [get]
func getMarket(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// parse inputs
		vars := mux.Vars(r)
		id, err := dnTypes.NewIDFromString(vars[MarketID])
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("%q param parsing: %v", MarketID, err))
			return
		}

		// prepare request
		req := types.MarketReq{
			ID: id,
		}

		bz, err := cliCtx.Codec.MarshalJSON(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		// query and parse the result
		res, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/market", types.ModuleName), bz)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		rest.PostProcessResponse(w, cliCtx, res)
	}
}
