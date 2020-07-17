package swagger

import (
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

var SwaggerInfo = swaggerInfo{
	Version:     "3.0",
	Host:        "127.0.0.1:1317",
	BasePath:    "/",
	Schemes:     []string{"http", "https"},
	Title:       "Dfinance dnode REST API",
	Description: "Dfinance API",
}

func RegisterRESTRoute(r *mux.Router) {
	r.PathPrefix("/swagger-ui/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger-ui/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))

	cors := handlers.CORS(
		handlers.AllowedHeaders([]string{"content-type"}),
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowCredentials(),
	)

	r.Use(cors)
}
