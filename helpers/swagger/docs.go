package swagger

import (
	"encoding/json"
	"log"

	"github.com/dfinance/dnode/cmd/dncli/docs"
	"github.com/ghodss/yaml"
	"github.com/swaggo/swag"
)

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

var doc = ""

type s struct{}

func (s *s) ReadDoc() string {
	return doc
}

func init() {
	defer swag.Register(swag.Name, &s{})

	var swagStruct map[string]interface{}

	err := yaml.Unmarshal([]byte(docs.Swagger), &swagStruct)
	if err != nil {
		log.Printf("Swagger unmarshal error: %v", err)
		return
	}

	swagStruct["host"] = SwaggerInfo.Host
	swagStruct["basePath"] = SwaggerInfo.BasePath
	swagStruct["schemes"] = SwaggerInfo.Schemes
	swagStruct["info"].(map[string]interface{})["version"] = SwaggerInfo.Version
	swagStruct["info"].(map[string]interface{})["title"] = SwaggerInfo.Title
	swagStruct["info"].(map[string]interface{})["description"] = SwaggerInfo.Description

	m, _ := json.Marshal(swagStruct)

	doc = string(m)
}
