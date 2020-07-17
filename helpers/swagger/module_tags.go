package swagger

var moduleTags = []swaggerTag{
	{
		Name:        "Currencies",
		Description: "Currencies DN module APIs: currency, coins issue/withdraw info",
	},
	{
		Name:        "Multisig",
		Description: "Multi signature DN module APIs: multisig calls info",
	},
	{
		Name:        "PoA",
		Description: "Proof of Authority DN module APIs: multisig validators info",
	},
	{
		Name:        "Oracle",
		Description: "Oracle DN module APIs: asset oracle prices",
	},
	{
		Name:        "Markets",
		Description: "Markets DN DEX module APIs: markets info",
	},
	{
		Name:        "Orders",
		Description: "Orders DN DEX module APIs: orders info",
	},
	{
		Name:        "VM",
		Description: "Virtual machine DN module APIs: DVM integration",
	},
}
