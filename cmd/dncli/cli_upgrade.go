package main

import (
	sdkClient "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeCli "github.com/cosmos/cosmos-sdk/x/upgrade/client/cli"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
)

func GetUpgradeTxCmd(cdc *amino.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:   upgrade.ModuleName,
		Short: "Upgrade transactions subcommands",
	}

	txCmd.AddCommand(sdkClient.PostCommands(
		upgradeCli.GetCmdSubmitUpgradeProposal(cdc),
		upgradeCli.GetCmdSubmitCancelUpgradeProposal(cdc),
	)...)

	return txCmd
}
