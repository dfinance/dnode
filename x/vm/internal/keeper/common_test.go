package keeper

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
	docker "github.com/fsouza/go-dockerclient"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	vmConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/vm/internal/types"
	"github.com/dfinance/dnode/x/vmauth"
)

const (
	accountHex = "2eb8d97a078f3ae572b0ea70362080c3e188a7e6000000000000000000000000"
	moveCode   = "a11ceb0b01000b016a00000004000000026e0000000800000003760000000c0000000b82000000060000000c88000000210000000da90000002500000005ce000000620000000430010000400000000870010000040000000974010000060000000a7a0100006d00000000000101000201000102010000030000040100050200060301080100010502000208010005000201080000010500020108000000000201080100010800000003030801000508000003000304050800000608000005030208000005030308000008010005124561726d61726b65644c69627261436f696e094c69627261436f696e01540663726561746513636c61696d5f666f725f726563697069656e7411636c61696d5f666f725f63726561746f7206756e7772617004636f696e09726563697069656e742eb8d97a078f3ae572b0ea70362080c3e188a7e6000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020200000700000801000100020007000b000b011300010c020b023100010201010100020212000b003000010c010e010c022c0c030b021001150b032221041000066300000000000000280b010202010100010307002c0c010b013000010c000b0002030100010406000b001400010c020c010b0102"
	movePath   = "00070b2b1ef472990ed03aa068408da8905c5a176639db1d35dc496d4f70c3c94a"
	value      = "68656c6c6f2c20776f726c6421"

	DefaultMockVMAddress        = "127.0.0.1:60051" // Default virtual machine address to connect from Cosmos SDK.
	DefaultMockDataListen       = "127.0.0.1:60052" // Default data server address to listen for connections from VM.
	DefaultMockVMTimeoutDeploy  = 100               // Default timeout for deploy module request.
	DefaultMockVMTimeoutExecute = 100               // Default timeout for execute request.

	FlagVMMockAddress = "vm.mock.address"
	FlagDSMockListen  = "ds.mock.listen"

	FlagVMAddress = "vm.address"
	FlagDSListen  = "ds.listen"
	FlagCompiler  = "vm.compiler"
)

type VMServer struct {
	vm_grpc.UnimplementedVMServiceServer
}

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context

	k  Keeper
	ak vmauth.VMAccountKeeper
	pk params.Keeper
	vk Keeper
	ok oracle.Keeper

	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keyOracle  *sdk.KVStoreKey
	keyParams  *sdk.KVStoreKey
	tkeyParams *sdk.TransientStoreKey
	keyVM      *sdk.KVStoreKey

	pathBytes    []byte
	codeBytes    []byte
	addressBytes []byte
	valueBytes   []byte

	//rawServer   *grpc.Server
	rawVMServer *grpc.Server
	vmServer    *VMServer

	dsListener *bufconn.Listener
	dsPort     int
}

var (
	vmMockAddress  *string
	dataListenMock *string

	vmAddress  *string
	vmCompiler *string
	dsListen   *string

	bufferSize = 1024 * 1024
)

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

func MockVMConfig() *vmConfig.VMConfig {
	return &vmConfig.VMConfig{
		Address:        *vmMockAddress,
		DataListen:     *dataListenMock,
		TimeoutDeploy:  DefaultMockVMTimeoutDeploy,
		TimeoutExecute: DefaultMockVMTimeoutExecute,
	}
}

func VMConfigWithFlags() *vmConfig.VMConfig {
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

func closeInput(_ testInput) {
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

func setupTestInput(launchMock bool) testInput {
	input := testInput{
		cdc:        codec.New(),
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey("acc"),
		//keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyOracle:  sdk.NewKVStoreKey("oracle"),
		keyParams:  sdk.NewKVStoreKey("params"),
		tkeyParams: sdk.NewTransientStoreKey("transient_params"),
		keyVM:      sdk.NewKVStoreKey("vm"),
	}

	types.RegisterCodec(input.cdc)
	auth.RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)
	oracle.RegisterCodec(input.cdc)

	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyMain, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyOracle, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	mstore.MountStoreWithDB(input.keyVM, sdk.StoreTypeIAVL, db)
	err := mstore.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	var vmListener *bufconn.Listener
	if launchMock {
		input.vmServer, input.rawVMServer, vmListener = LaunchVMMock()
	}

	var config *vmConfig.VMConfig

	if launchMock {
		config = MockVMConfig()
	} else {
		config = VMConfigWithFlags()
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
		clientConn, err = grpc.DialContext(ctx, "", grpc.WithContextDialer(GetBufDialer(vmListener)), grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
	} else {
		clientConn, err = grpc.Dial(config.Address, grpc.WithInsecure())
		if err != nil {
			panic(err)
		}
	}

	input.vk = Keeper{
		cdc:      input.cdc,
		storeKey: input.keyVM,
		client:   vm_grpc.NewVMServiceClient(clientConn),
		listener: listener,
		config:   config,
	}

	input.pk = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams, params.DefaultCodespace)
	input.ak = vmauth.NewVMAccountKeeper(
		input.cdc,
		input.keyAccount,
		input.pk.Subspace(auth.DefaultParamspace),
		input.vk,
		auth.ProtoBaseAccount,
	)

	input.ok = oracle.NewKeeper(input.keyOracle, input.cdc, input.pk.Subspace(oracle.DefaultParamspace), oracle.DefaultCodespace, input.vk)

	input.vk.dsServer = NewDSServer(&input.vk)
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "dn-testnet-vm-keeper-test"}, false, log.NewNopLogger())

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

func (server VMServer) ExecuteContracts(ctx context.Context, req *vm_grpc.VMExecuteRequest) (*vm_grpc.VMExecuteResponses, error) {
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

func LaunchVMMock() (*VMServer, *grpc.Server, *bufconn.Listener) {
	vmListener := bufconn.Listen(bufferSize)

	vmServer := VMServer{}
	server := grpc.NewServer()
	vm_grpc.RegisterVMServiceServer(server, vmServer)

	go func() {
		if err := server.Serve(vmListener); err != nil {
			panic(err)
		}
	}()

	return &vmServer, server, vmListener
}

func GetBufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return listener.Dial()
	}
}

func createVMOptions(registry, dsServerUrl, tag string) docker.CreateContainerOptions {
	ports := make(map[docker.Port]struct{})
	ports["50051/tcp"] = struct{}{}

	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        registry + "/dfinance/dvm:" + tag,
			ExposedPorts: ports,
			Cmd: []string{
				"./dvm",
				"0.0.0.0:50051",
				dsServerUrl,
			},
		},
		HostConfig: &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				"50051/tcp": {{HostIP: "0.0.0.0", HostPort: "50051"}},
			},
		},
	}

	return opts
}

// creating compiler options.
func createCompilerOptions(registry, dsServerUrl, tag string) docker.CreateContainerOptions {
	ports := make(map[docker.Port]struct{})
	ports["50053/tcp"] = struct{}{}

	opts := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:        registry + "/dfinance/dvm:" + tag,
			ExposedPorts: ports,
			Cmd: []string{
				"./compiler",
				"0.0.0.0:50053",
				dsServerUrl,
			},
		},
		HostConfig: &docker.HostConfig{
			PortBindings: map[docker.Port][]docker.PortBinding{
				"50053/tcp": {{HostIP: "0.0.0.0", HostPort: "50053"}},
			},
		},
	}

	return opts
}

// stop docker
func stopDocker(t *testing.T, client *docker.Client, container *docker.Container) {
	if err := client.RemoveContainer(docker.RemoveContainerOptions{
		ID:    container.ID,
		Force: true,
	}); err != nil {
		t.Fatalf("can't remove container: %v", err)
	}
}

// Launch docker container with dvm.
func launchDocker(dsServerUrl string, t *testing.T) (*docker.Client, *docker.Container, *docker.Container) {
	tag := os.Getenv("TAG")
	if tag == "" {
		tag = "master"
	}

	registry := os.Getenv("REGISTRY")
	if registry == "" {
		t.Fatalf("provide REGISTRY via env, e.g. REGISTRY=...")
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		t.Fatalf("can't connect to docker: %v", err)
	}

	compiler, err := client.CreateContainer(createCompilerOptions(registry, dsServerUrl, tag))
	if err != nil {
		t.Fatalf("can't create container for compiler: %v", err)
	}

	err = client.StartContainer(compiler.ID, nil)
	if err != nil {
		t.Fatalf("cannot start docker container for compiler: %v", err)
	}

	vm, err := client.CreateContainer(createVMOptions(registry, dsServerUrl, tag))
	if err != nil {
		t.Fatalf("can't create container for vm: %v", err)
	}

	err = client.StartContainer(vm.ID, nil)
	if err != nil {
		t.Fatalf("can't start docker container for vm: %v", err)
	}

	return client, compiler, vm
}

// waitStarted waits for a container to start for the maxWait time.
func waitStarted(client *docker.Client, id string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := client.InspectContainerWithOptions(docker.InspectContainerOptions{
			ID: id,
		})
		if err != nil {
			break
		}
		if c.State.Running {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot start container %s for %v", id, maxWait)
}

// waitReachable waits for hostport to became reachable for the maxWait time.
func waitReachable(hostport string, maxWait time.Duration) error {
	done := time.Now().Add(maxWait)
	for time.Now().Before(done) {
		c, err := net.Dial("tcp", hostport)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("cannot connect %v for %v", hostport, maxWait)
}
