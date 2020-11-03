module github.com/dfinance/dnode

go 1.14

replace github.com/cosmos/cosmos-sdk => github.com/dfinance/cosmos-sdk v0.39.2-0.20201103114952-6e10e3e0ab85

// Local development option
//replace github.com/cosmos/cosmos-sdk => /Users/boris/go/src/github.com/dfinance/cosmos-sdk
//replace github.com/cosmos/cosmos-sdk => /Users/tiky/Go_Projects/src/github.com/dfinance/cosmos-sdk

// Fix of OS X hostmachine test runs
// Source: https://github.com/ory/dockertest/issues/208
replace golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6

require (
	github.com/99designs/keyring v1.1.3
	github.com/OneOfOne/xxhash v1.2.7
	github.com/atlassian/go-sentry-api v0.0.0-20200117001222-a9ccec16c98b
	github.com/cosmos/cosmos-sdk v0.0.1
	github.com/dfinance/dvm-proto/go v0.0.0-20200819065410-6b70956c85de
	github.com/dfinance/glav v0.0.0-20200814081332-c4701f6c12a6
	github.com/dfinance/lcs v0.1.7-big
	github.com/fsouza/go-dockerclient v1.6.6-0.20200910033347-214a51d9a1e5
	github.com/getsentry/sentry-go v0.5.1
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/gorilla/handlers v1.4.2
	github.com/gorilla/mux v1.8.0
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/olekukonko/tablewriter v0.0.4
	github.com/pelletier/go-toml v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/rakyll/statik v0.1.7 // indirect
	github.com/shopspring/decimal v1.2.0
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.6.1
	github.com/swaggo/http-swagger v0.0.0-20200308142732-58ac5e232fba
	github.com/swaggo/swag v1.6.7
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.7
	github.com/tendermint/tm-db v0.5.1
	google.golang.org/grpc v1.30.0
	google.golang.org/protobuf v1.24.0 // indirect
	k8s.io/apimachinery v0.18.6 // indirect
	k8s.io/kubernetes v1.13.0
)
