package config

import (
	"text/template"
)

var configTemplate *template.Template

func init() {
	var err error
	tmpl := template.New("vmConfigTemplate")

	if configTemplate, err = tmpl.Parse(defaultConfigTemplate); err != nil {
		panic(err)
	}
}

const defaultConfigTemplate = `# This is a TOML config file to configurate connection to VM.
# For more information, see https://github.com/toml-lang/toml

##### main base config options #####

# VM network address to connect
vm_address = "{{ .Address }}"

# VM deploy request timeout in milliseconds
vm_deploy_timeout = {{ .DeployTimeout }}

# VM data server listen address
vm_data_listen = "{{ .DataListen }}"
`
