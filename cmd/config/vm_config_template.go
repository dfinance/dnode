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

# VM network address to connect.
vm_address = "{{ .Address }}"

# VM data server listen address.
vm_data_listen = "{{ .DataListen }}

# VM retry settings.

## Retry max attempts.
## Default is 0 - infinity attempts, -1 - to desable.
vm_retry_max_attempts = {{ .MaxAttempts }}

## Initial backoff in ms.
## Default is 100ms.
vm_retry_initial_backoff = {{ .InitialBackoff }}

## Max backoff in ms.
## Default is 150ms.
vm_retry_max_backoff = {{ .MaxBackoff }}

## Backoff multiplier.
## Default 
vm_retry_backoff_multiplier = {{ .BackoffMultiplier }}
`
