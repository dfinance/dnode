package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdkSrvConfig "github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilTypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmConfig "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	tmOs "github.com/tendermint/tendermint/libs/os"
	tmRand "github.com/tendermint/tendermint/libs/rand"
	tmTypes "github.com/tendermint/tendermint/types"
	tmTime "github.com/tendermint/tendermint/types/time"

	dnConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/genaccounts"
	"github.com/dfinance/dnode/x/oracle"
)

// DONTCOVER

var (
	flagNodeDirPrefix          = "node-dir-prefix"
	flagNumValidators          = "v"
	flagOutputDir              = "output-dir"
	flagNodeDaemonHome         = "node-daemon-home"
	flagNodeCLIHome            = "node-cli-home"
	flagStartingIPAddress      = "starting-ip-address"
	flagComissionRate          = "comission-rate"
	flagComissionMaxRate       = "comission-max-rate"
	flagComissionMaxChangeRate = "comission-max-change-rate"
)

type cliFlags struct {
	outputDir              string
	chainID                string
	minGasPrices           string
	nodeDirPrefix          string
	nodeDaemonHome         string
	nodeCLIHome            string
	startingIPAddress      string
	numValidators          int
	comissionRate          sdk.Dec
	comissionMaxRate       sdk.Dec
	comissionMaxChangeRate sdk.Dec
}

// get cmd to initialize all files for tendermint testnet and application
func testnetCmd(ctx *server.Context, cdc *codec.Codec, mbm module.BasicManager, genAccIterator genutilTypes.GenesisAccountsIterator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a Dnode testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:
	dnode testnet --v 4 --output-dir ./output --starting-ip-address 192.168.10.2
	`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			config := ctx.Config

			cf := cliFlags{
				outputDir:         viper.GetString(flagOutputDir),
				chainID:           viper.GetString(flags.FlagChainID),
				minGasPrices:      viper.GetString(server.FlagMinGasPrices),
				nodeDirPrefix:     viper.GetString(flagNodeDirPrefix),
				nodeDaemonHome:    viper.GetString(flagNodeDaemonHome),
				nodeCLIHome:       viper.GetString(flagNodeCLIHome),
				startingIPAddress: viper.GetString(flagStartingIPAddress),
				numValidators:     viper.GetInt(flagNumValidators),
			}
			var err error
			cf.comissionRate, err = sdk.NewDecFromStr(viper.GetString(flagComissionRate))
			if err != nil {
				return err
			}
			cf.comissionMaxRate, err = sdk.NewDecFromStr(viper.GetString(flagComissionMaxRate))
			if err != nil {
				return err
			}
			cf.comissionMaxChangeRate, err = sdk.NewDecFromStr(viper.GetString(flagComissionMaxChangeRate))
			if err != nil {
				return err
			}

			return InitTestnet(cmd, config, cdc, mbm, genAccIterator, &cf)
		},
	}

	cmd.Flags().Int(flagNumValidators, 4,
		"Number of validators to initialize the testnet with")
	cmd.Flags().StringP(flagOutputDir, "o", "./mytestnet",
		"Directory to store initialization data for the testnet")
	cmd.Flags().String(flagNodeDirPrefix, "node",
		"Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(flagNodeDaemonHome, "dnode",
		"Home directory of the node's daemon configuration")
	cmd.Flags().String(flagNodeCLIHome, "dncli",
		"Home directory of the node's cli configuration")
	cmd.Flags().String(flagStartingIPAddress, "192.168.0.1",
		"Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().String(flags.FlagChainID, "",
		"genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().String(server.FlagMinGasPrices, fmt.Sprintf("1%s", dnConfig.MainDenom),
		"Minimum gas prices to accept for transactions; All fees in a tx must meet this minimum (e.g. 0.01photino,0.001stake)")
	cmd.Flags().String(flagComissionRate, "0.100000000000000000",
		"Comission rate")
	cmd.Flags().String(flagComissionMaxRate, "0.200000000000000000",
		"Comission rate")
	cmd.Flags().String(flagComissionMaxChangeRate, "0.010000000000000000",
		"Comission max change rate")
	//cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend,
	//	"Select keyring's backend (os|file|test)")

	return cmd
}

const nodeDirPerm = 0755

// Initialize the testnet
func InitTestnet(cmd *cobra.Command, config *tmConfig.Config, cdc *codec.Codec, mbm module.BasicManager, genAccIterator genutilTypes.GenesisAccountsIterator, cf *cliFlags) error {
	if cf.chainID == "" {
		cf.chainID = "chain-" + tmRand.Str(6)
	}

	monikers := make([]string, cf.numValidators)
	nodeIDs := make([]string, cf.numValidators)
	valPubKeys := make([]crypto.PubKey, cf.numValidators)

	dnCfg := sdkSrvConfig.DefaultConfig()
	dnCfg.MinGasPrices = cf.minGasPrices

	// nolint:prealloc
	var (
		genAccounts genaccounts.GenesisState
		genFiles    []string
	)

	// generate private keys, node IDs, and initial transactions
	for i := 0; i < cf.numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", cf.nodeDirPrefix, i)
		nodeDir := path.Join(cf.outputDir, nodeDirName, cf.nodeDaemonHome)
		clientDir := path.Join(cf.outputDir, nodeDirName, cf.nodeCLIHome)
		gentxsDir := path.Join(cf.outputDir, "gentxs")

		config.SetRoot(nodeDir)
		config.RPC.ListenAddress = "tcp://0.0.0.0:26657"

		if err := os.MkdirAll(path.Join(nodeDir, "config"), nodeDirPerm); err != nil {
			_ = os.RemoveAll(cf.outputDir)
			return err
		}

		if err := os.MkdirAll(clientDir, nodeDirPerm); err != nil {
			_ = os.RemoveAll(cf.outputDir)
			return err
		}

		monikers = append(monikers, nodeDirName)
		config.Moniker = nodeDirName

		ip, err := getIP(i, cf.startingIPAddress)
		if err != nil {
			_ = os.RemoveAll(cf.outputDir)
			return err
		}

		nodeIDs[i], valPubKeys[i], err = genutil.InitializeNodeValidatorFiles(config)
		if err != nil {
			_ = os.RemoveAll(cf.outputDir)
			return err
		}

		memo := fmt.Sprintf("%s@%s:26656", nodeIDs[i], ip)
		genFiles = append(genFiles, config.GenesisFile())

		kb, err := keys.NewKeyBaseFromDir(clientDir)
		if err != nil {
			return err
		}

		keyPass := keys.DefaultKeyPass
		addr, secret, err := server.GenerateSaveCoinKey(kb, nodeDirName, keyPass, true)
		if err != nil {
			_ = os.RemoveAll(cf.outputDir)
			return err
		}

		info := map[string]string{"secret": secret}

		cliPrint, err := json.Marshal(info)
		if err != nil {
			return err
		}

		// save private key seed words
		if err := writeFile(fmt.Sprintf("%v.json", "key_seed"), clientDir, cliPrint); err != nil {
			return err
		}

		accStakingTokens := sdk.TokensFromConsensusPower(500000000000)
		coins := sdk.Coins{
			sdk.NewCoin(dnConfig.MainDenom, accStakingTokens),
		}

		genAccounts = append(genAccounts, *auth.NewBaseAccount(addr, coins.Sort(), nil, 0, 0))

		valTokens := sdk.TokensFromConsensusPower(500000)
		msg := staking.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			valPubKeys[i],
			sdk.NewCoin(dnConfig.MainDenom, valTokens),
			staking.NewDescription(nodeDirName, "", "", "", ""),
			staking.NewCommissionRates(cf.comissionRate, cf.comissionMaxRate, cf.comissionMaxChangeRate),
			sdk.OneInt(),
		)

		inBuf := bufio.NewReader(cmd.InOrStdin())
		txBldr := auth.NewTxBuilderFromCLI(inBuf).WithChainID(cf.chainID).WithMemo(memo).WithKeybase(kb)
		tx := auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{Gas: 200000}, []auth.StdSignature{}, memo)

		signedTx, err := txBldr.SignStdTx(nodeDirName, keys.DefaultKeyPass, tx, false)
		if err != nil {
			_ = os.RemoveAll(cf.outputDir)
			return err
		}

		txBytes, err := cdc.MarshalJSON(signedTx)
		if err != nil {
			_ = os.RemoveAll(cf.outputDir)
			return err
		}

		// gather gentxs folder
		if err := writeFile(fmt.Sprintf("%v.json", nodeDirName), gentxsDir, txBytes); err != nil {
			_ = os.RemoveAll(cf.outputDir)
			return err
		}

		// TODO: Rename config file to server.toml as it's not particular to Dn
		// (REF: https://github.com/cosmos/cosmos-sdk/issues/4125).
		dnConfigpath := path.Join(nodeDir, "config/app.toml")
		sdkSrvConfig.WriteConfigFile(dnConfigpath, dnCfg)
	}

	if err := initGenFiles(cdc, mbm, cf.chainID, genAccounts, genFiles, cf.numValidators); err != nil {
		return err
	}

	err := collectGenFiles(
		cdc, config, cf.chainID, monikers, nodeIDs, valPubKeys, cf.numValidators,
		cf.outputDir, cf.nodeDirPrefix, cf.nodeDaemonHome, genAccIterator,
	)
	if err != nil {
		return err
	}

	cmd.PrintErrf("Successfully initialized %d node directories\n", cf.numValidators)
	return nil
}

func initGenFiles(
	cdc *codec.Codec, mbm module.BasicManager, chainID string,
	genAccounts genaccounts.GenesisState, genFiles []string, numValidators int,
) error {

	appGenState := mbm.DefaultGenesis()

	// set the accounts in the genesis state
	appGenState[genaccounts.ModuleName] = cdc.MustMarshalJSON(genAccounts)

	stakingDataBz := appGenState[staking.ModuleName]
	var stakingGenState staking.GenesisState
	cdc.MustUnmarshalJSON(stakingDataBz, &stakingGenState)
	stakingGenState.Params.BondDenom = dnConfig.MainDenom
	appGenState[staking.ModuleName] = cdc.MustMarshalJSON(stakingGenState)

	oracleDataBz := appGenState[oracle.ModuleName]
	var oracleGenState oracle.GenesisState
	cdc.MustUnmarshalJSON(oracleDataBz, &oracleGenState)
	nomenees := make([]string, len(genAccounts))
	for i, acc := range genAccounts {
		nomenees[i] = acc.Address.String()
	}
	oracleGenState.Params.Nominees = nomenees
	appGenState[oracle.ModuleName] = cdc.MustMarshalJSON(oracleGenState)

	appGenStateJSON, err := codec.MarshalJSONIndent(cdc, appGenState)
	if err != nil {
		return err
	}

	genDoc := tmTypes.GenesisDoc{
		ChainID:    chainID,
		AppState:   appGenStateJSON,
		Validators: nil,
	}

	// generate empty genesis files for each validator and save
	for i := 0; i < numValidators; i++ {
		if err := genDoc.SaveAs(genFiles[i]); err != nil {
			return err
		}
	}
	return nil
}

func collectGenFiles(
	cdc *codec.Codec, config *tmConfig.Config, chainID string,
	monikers, nodeIDs []string, valPubKeys []crypto.PubKey,
	numValidators int, outputDir, nodeDirPrefix, nodeDaemonHome string,
	genAccIterator genutilTypes.GenesisAccountsIterator) error {

	var appState json.RawMessage
	genTime := tmTime.Now()

	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", nodeDirPrefix, i)
		nodeDir := path.Join(outputDir, nodeDirName, nodeDaemonHome)
		gentxsDir := path.Join(outputDir, "gentxs")
		moniker := monikers[i]
		config.Moniker = nodeDirName

		config.SetRoot(nodeDir)

		nodeID, valPubKey := nodeIDs[i], valPubKeys[i]
		initCfg := genutil.NewInitConfig(chainID, gentxsDir, moniker, nodeID, valPubKey)

		genDoc, err := tmTypes.GenesisDocFromFile(config.GenesisFile())
		if err != nil {
			return err
		}

		nodeAppState, err := genutil.GenAppStateFromConfig(cdc, config, initCfg, *genDoc, genAccIterator)
		if err != nil {
			return err
		}

		if appState == nil {
			// set the canonical application state (they should not differ)
			appState = nodeAppState
		}

		genFile := config.GenesisFile()

		// overwrite each validator's genesis file to have a canonical genesis time
		if err := genutil.ExportGenesisFileWithTime(genFile, chainID, nil, appState, genTime); err != nil {
			return err
		}
	}

	return nil
}

func getIP(i int, startingIPAddr string) (ip string, err error) {
	if len(startingIPAddr) == 0 {
		ip, err = server.ExternalIP()
		if err != nil {
			return "", err
		}
		return ip, nil
	}
	return calculateIP(startingIPAddr, i)
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}

	return ipv4.String(), nil
}

func writeFile(name string, dir string, contents []byte) error {
	writePath := path.Join(dir)
	file := path.Join(writePath, name)

	if err := tmOs.EnsureDir(writePath, 0700); err != nil {
		return err
	}

	if err := tmOs.WriteFile(file, contents, 0600); err != nil {
		return err
	}

	return nil
}
