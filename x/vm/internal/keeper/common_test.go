// +build unit integ

package keeper

import (
	"context"
	"encoding/hex"
	"flag"
	"math/rand"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	cfg "github.com/tendermint/tendermint/config"
	tmflags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	vmConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/helpers/tests"
	"github.com/dfinance/dnode/x/currencies_register"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/vm/internal/types"
	"github.com/dfinance/dnode/x/vmauth"
)

const (
	accountHex = "2eb8d97a078f3ae572b0ea70362080c3e188a7e6000000000000000000000000"
	moveCode   = "a11ceb0b01000b016a00000004000000026e0000000800000003760000000c0000000b82000000060000000c88000000210000000da90000002500000005ce000000620000000430010000400000000870010000040000000974010000060000000a7a0100006d00000000000101000201000102010000030000040100050200060301080100010502000208010005000201080000010500020108000000000201080100010800000003030801000508000003000304050800000608000005030208000005030308000008010005124561726d61726b65644c69627261436f696e094c69627261436f696e01540663726561746513636c61696d5f666f725f726563697069656e7411636c61696d5f666f725f63726561746f7206756e7772617004636f696e09726563697069656e742eb8d97a078f3ae572b0ea70362080c3e188a7e6000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020200000700000801000100020007000b000b011300010c020b023100010201010100020212000b003000010c010e010c022c0c030b021001150b032221041000066300000000000000280b010202010100010307002c0c010b013000010c000b0002030100010406000b001400010c020c010b0102"
	movePath   = "00070b2b1ef472990ed03aa068408da8905c5a176639db1d35dc496d4f70c3c94a"
	value      = "68656c6c6f2c20776f726c6421"

	DefaultMockVMAddress  = "127.0.0.1:60051" // Default virtual machine address to connect from Cosmos SDK.
	DefaultMockDataListen = "127.0.0.1:60052" // Default data server address to listen for connections from VM.
	FlagVMMockAddress     = "vm.mock.address"
	FlagDSMockListen      = "ds.mock.listen"

	FlagVMAddress = "vm.address"
	FlagDSListen  = "ds.listen"
	FlagCompiler  = "vm.compiler"

	CoinsInfo = `{ 
		"currencies": [
			{
				"path": "0172c9f1bfe0a2bf6ac342aaa3c3380852d4694ae4e71655d37aa5d2e6700ed94e",
          		"denom": "dfi",
          		"decimals": 18,
          		"totalSupply": "100000000000000000000000000"
        	},
        	{
          		"path": "0116fbac6fd286d2bfec4549161245982b730291a9cbc5281f5fcfb41aeb7bfb26",
          		"denom": "eth",
          		"decimals": 18,
          		"totalSupply": "100000000000000000000000000"
        	},
			{
          		"path": "01e10f377b920a0a8a330edd7beff6c3a11cdeb7682c964b02aa5bb6a784b84920",
          		"denom": "usdt",
          		"decimals": 6,
          		"totalSupply": "10000000000000"
        	},
        	{
          		"path": "0158c690830c7e2f25b85de6ab85052fd1e79e6a9cbb52a9740be7ff7275604c1b",
          		"denom": "btc",
          		"decimals": 8,
          		"totalSupply": "100000000000000"
			}
		]
	}`
)

var (
	vmMockAddress  *string
	dataListenMock *string

	vmAddress  *string
	vmCompiler *string
	dsListen   *string

	bufferSize = 1024 * 1024
)

type vmServer struct {
	vm_grpc.UnimplementedVMServiceServer
}

func (server vmServer) ExecuteContracts(ctx context.Context, req *vm_grpc.VMExecuteRequest) (*vm_grpc.VMExecuteResponses, error) {
	// execute module
	resps := &vm_grpc.VMExecuteResponses{
		Executions: make([]*vm_grpc.VMExecuteResponse, len(req.Contracts)),
	}

	for i, contract := range req.Contracts {
		if contract.ContractType == vm_grpc.ContractType_Module {
			// process module
			values := make([]*vm_grpc.VMValue, 1)
			values[0] = &vm_grpc.VMValue{
				Type:  vm_grpc.VmWriteOp_Value,
				Value: randomValue(512),
				Path:  randomPath(),
			}

			resps.Executions[i] = &vm_grpc.VMExecuteResponse{
				WriteSet: values,
				Events:   nil,
				GasUsed:  10000,
				Status:   vm_grpc.ContractStatus_Keep,
			}
		} else if contract.ContractType == vm_grpc.ContractType_Script {
			// process script
			values := make([]*vm_grpc.VMValue, 2)
			values[0] = &vm_grpc.VMValue{
				Type:  vm_grpc.VmWriteOp_Value,
				Value: randomValue(8),
				Path:  randomPath(),
			}
			values[1] = &vm_grpc.VMValue{
				Type:  vm_grpc.VmWriteOp_Value,
				Value: randomValue(32),
				Path:  randomPath(),
			}

			events := make([]*vm_grpc.VMEvent, 1)
			events[0] = &vm_grpc.VMEvent{
				Key:            []byte("test event"),
				SequenceNumber: 0,
				Type: &vm_grpc.VMType{
					Tag: vm_grpc.VMTypeTag_ByteArray,
				},
				EventData: randomValue(32),
			}

			resps.Executions[i] = &vm_grpc.VMExecuteResponse{
				WriteSet: values,
				Events:   events,
				GasUsed:  10000,
				Status:   vm_grpc.ContractStatus_Keep,
			}
		} else {
			panic("wrong contract type")
		}
	}

	return resps, nil
}

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context

	ak vmauth.VMAccountKeeper
	pk params.Keeper
	vk Keeper
	ok oracle.Keeper
	cr currencies_register.Keeper

	keyMain      *sdk.KVStoreKey
	keyAccount   *sdk.KVStoreKey
	keyOracle    *sdk.KVStoreKey
	keyCRegister *sdk.KVStoreKey
	keyParams    *sdk.KVStoreKey
	tkeyParams   *sdk.TransientStoreKey
	keyVM        *sdk.KVStoreKey

	pathBytes    []byte
	codeBytes    []byte
	addressBytes []byte
	valueBytes   []byte

	rawVMServer *grpc.Server
	vmServer    *vmServer

	dsListener *bufconn.Listener
	dsPort     int
}

func (i *testInput) Stop() {
	// go func() {
	// 	if input.rawServer != nil {
	// 		input.rawServer.GracefulStop()
	// 	}
	//
	// 	if input.rawVMServer != nil {
	// 		input.rawVMServer.GracefulStop()
	// 	}
	//
	// 	input.vk.listener.Close()
	//
	// 	if input.dsListener != nil {
	// 		input.dsListener.Close()
	// 	}
	// }()
}

func newTestInput(launchMock bool) testInput {
	input := testInput{
		cdc:        codec.New(),
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey("acc"),
		//keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyOracle:    sdk.NewKVStoreKey("oracle"),
		keyParams:    sdk.NewKVStoreKey("params"),
		keyCRegister: sdk.NewKVStoreKey("coins_register"),
		tkeyParams:   sdk.NewTransientStoreKey("transient_params"),
		keyVM:        sdk.NewKVStoreKey("vm"),
	}

	types.RegisterCodec(input.cdc)
	auth.RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)
	oracle.RegisterCodec(input.cdc)
	gov.RegisterCodec(input.cdc)

	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyMain, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyOracle, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCRegister, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	mstore.MountStoreWithDB(input.keyVM, sdk.StoreTypeIAVL, db)
	err := mstore.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	var vmListener *bufconn.Listener
	if launchMock {
		input.vmServer, input.rawVMServer, vmListener = startMockDSServer()
	}

	var config *vmConfig.VMConfig
	if launchMock {
		config = newMockVMConfig()
	} else {
		config = newVMConfigWithFlags()
	}

	// process if mock
	var listener net.Listener
	if launchMock {
		input.dsListener = bufconn.Listen(bufferSize)
		listener = input.dsListener
	} else {
		listener, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		input.dsPort = listener.Addr().(*net.TCPAddr).Port
	}

	// no blocking, so we can launch part of tests even without vm
	var clientConn *grpc.ClientConn
	if launchMock {
		ctx := context.TODO()
		clientConn, err = grpc.DialContext(ctx, "", grpc.WithContextDialer(getBufDialer(vmListener)), grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
	} else {
		clientConn, err = helpers.GetGRpcClientConnection(config.Address, 1*time.Second)
		if err != nil {
			panic(err)
		}
	}

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	logger, err = tmflags.ParseLogLevel("x/vm:info,x/vm/dsserver:info", logger, cfg.DefaultLogLevel())
	if err != nil {
		panic(err)
	}
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "dn-testnet-vm-keeper-test"}, false, logger)

	input.vk = NewKeeper(input.keyVM, input.cdc, clientConn, listener, config)
	input.cr = currencies_register.NewKeeper(input.cdc, input.keyCRegister, input.vk)
	input.pk = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams)
	input.ak = vmauth.NewVMAccountKeeper(input.cdc, input.keyAccount, input.pk.Subspace(auth.DefaultParamspace), input.vk, auth.ProtoBaseAccount)
	input.ok = oracle.NewKeeper(input.keyOracle, input.cdc, input.pk.Subspace(oracle.DefaultParamspace), input.vk)

	err = input.cr.InitGenesis(input.ctx, []byte(CoinsInfo))
	if err != nil {
		panic(err)
	}

	input.addressBytes, err = hex.DecodeString(accountHex)
	if err != nil {
		panic(err)
	}

	input.pathBytes, err = hex.DecodeString(movePath)
	if err != nil {
		panic(err)
	}

	input.codeBytes, err = hex.DecodeString(moveCode)
	if err != nil {
		panic(err)
	}

	input.valueBytes, err = hex.DecodeString(value)
	if err != nil {
		panic(err)
	}

	return input
}

func newMockVMConfig() *vmConfig.VMConfig {
	return &vmConfig.VMConfig{
		Address:           *vmMockAddress,
		DataListen:        *dataListenMock,
		MaxAttempts:       vmConfig.DefaultMaxAttempts,
		InitialBackoff:    vmConfig.DefaultInitialBackoff,
		MaxBackoff:        vmConfig.DefaultMaxBackoff,
		BackoffMultiplier: vmConfig.DefaultBackoffMultiplier,
	}
}

func newVMConfigWithFlags() *vmConfig.VMConfig {
	config := vmConfig.DefaultVMConfig()
	config.Address = *vmAddress
	config.DataListen = *dsListen

	return config
}

func randomPath() *vm_grpc.VMAccessPath {
	return &vm_grpc.VMAccessPath{
		Address: randomValue(32),
		Path:    randomValue(20),
	}
}

func randomValue(len int) []byte {
	rndBytes := make([]byte, len)

	_, err := rand.Read(rndBytes)
	if err != nil {
		panic(err)
	}

	return rndBytes
}

func getBufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return listener.Dial()
	}
}

func startMockDSServer() (*vmServer, *grpc.Server, *bufconn.Listener) {
	vmListener := bufconn.Listen(bufferSize)

	vmServer := vmServer{}
	server := grpc.NewServer()
	vm_grpc.RegisterVMServiceServer(server, vmServer)

	go func() {
		if err := server.Serve(vmListener); err != nil {
			panic(err)
		}
	}()

	return &vmServer, server, vmListener
}

func startDVMContainer(t *testing.T, dsPort int) (stopFunc func()) {
	vmUrl, err := url.Parse(*vmAddress)
	require.NoError(t, err, "parsing vmAddress")

	containerStop := tests.LaunchDVMWithNetTransport(t, vmUrl.Port(), strconv.Itoa(dsPort), false)
	*vmCompiler = "127.0.0.1:" + vmUrl.Port()

	return containerStop
}

func init() {
	if flag.Lookup(FlagCompiler) == nil {
		vmCompiler = flag.String(FlagCompiler, "127.0.0.1:50053", "compiler address")
	}

	if flag.Lookup(FlagVMAddress) == nil {
		vmAddress = flag.String(FlagVMAddress, vmConfig.DefaultVMAddress, "Move VM address to connect during unit tests")
	}

	if flag.Lookup(FlagDSListen) == nil {
		dsListen = flag.String(FlagDSListen, vmConfig.DefaultDataListen, "address to listen of Data Server (DS) during unit tests")
	}

	if flag.Lookup(FlagVMMockAddress) == nil {
		vmMockAddress = flag.String(FlagVMMockAddress, DefaultMockVMAddress, "mocked address of virtual machine server client/server")
	}

	if flag.Lookup(FlagDSMockListen) == nil {
		dataListenMock = flag.String(FlagDSMockListen, DefaultMockDataListen, "address of mocked data server to launch/connect")
	}
}
