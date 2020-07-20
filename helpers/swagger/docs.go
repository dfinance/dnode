package swagger

import (
	"encoding/json"
	"log"

	"github.com/ghodss/yaml"
	"github.com/spf13/viper"
	"github.com/swaggo/swag"

	"github.com/dfinance/dnode/cmd/dncli/docs"
)

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
	Tags        []swaggerTag
}

type swaggerTag struct {
	Name        string
	Description string
}

var doc = ""

type s struct{}

func (s *s) ReadDoc() string {
	return doc
}

func Init() {
	defer swag.Register(swag.Name, &s{})

	// unmarshal merged Cosmos SDK and Dnode Swagger files
	var swagStruct map[string]interface{}
	err := yaml.Unmarshal([]byte(docs.Swagger), &swagStruct)
	if err != nil {
		log.Printf("Swagger YAML unmarshal error: %v", err)
		return
	}

	// overwrite fields
	swagStruct["host"] = viper.GetString("swagger-host")
	swagStruct["basePath"] = SwaggerInfo.BasePath
	swagStruct["schemes"] = viper.GetStringSlice("swagger-schemes")
	swagStruct["info"].(map[string]interface{})["version"] = SwaggerInfo.Version
	swagStruct["info"].(map[string]interface{})["title"] = SwaggerInfo.Title
	swagStruct["info"].(map[string]interface{})["description"] = SwaggerInfo.Description

	// append Swagger tags descriptions for Dnode modules
	swaggerTagObjs := swagStruct["tags"].([]interface{})
	for _, moduleTag := range moduleTags {
		swaggerTagObjs = append(swaggerTagObjs, map[string]interface{}{
			"name":        moduleTag.Name,
			"description": moduleTag.Description,
		})
	}
	swagStruct["tags"] = swaggerTagObjs

	// marshal the resulting Swagger JSON
	m, err := json.Marshal(swagStruct)
	if err != nil {
		log.Printf("Swagger JSON marshal error: %v", err)
		return
	}
	doc = string(m)
}
