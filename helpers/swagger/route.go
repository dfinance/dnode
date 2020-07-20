package swagger

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	httpSwagger "github.com/swaggo/http-swagger"
)

var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	BasePath:    "/",
	Title:       "Dfinance dnode REST API",
	Description: "Dfinance API",
}

func RegisterRESTRoute(r *mux.Router) {
	Init()

	r.PathPrefix("/swagger-ui/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger-ui/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins(viper.GetStringSlice("swagger-allowed-urls")),
		handlers.AllowCredentials(),
	)

	r.Use(cors)
}
