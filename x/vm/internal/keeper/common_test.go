package keeper

import (
	"encoding/hex"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"math/rand"
	"net"
	"time"
	vmConfig "wings-blockchain/cmd/config"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

const (
	accountHex = "2eb8d97a078f3ae572b0ea70362080c3e188a7e6000000000000000000000000"
	moveCode   = "a11ceb0b01000b016a00000004000000026e0000000800000003760000000c0000000b82000000060000000c88000000210000000da90000002500000005ce000000620000000430010000400000000870010000040000000974010000060000000a7a0100006d00000000000101000201000102010000030000040100050200060301080100010502000208010005000201080000010500020108000000000201080100010800000003030801000508000003000304050800000608000005030208000005030308000008010005124561726d61726b65644c69627261436f696e094c69627261436f696e01540663726561746513636c61696d5f666f725f726563697069656e7411636c61696d5f666f725f63726561746f7206756e7772617004636f696e09726563697069656e742eb8d97a078f3ae572b0ea70362080c3e188a7e6000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000020200000700000801000100020007000b000b011300010c020b023100010201010100020212000b003000010c010e010c022c0c030b021001150b032221041000066300000000000000280b010202010100010307002c0c010b013000010c000b0002030100010406000b001400010c020c010b0102"
	movePath   = "00070b2b1ef472990ed03aa068408da8905c5a176639db1d35dc496d4f70c3c94a"
	value      = "68656c6c6f2c20776f726c6421"
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context

	k  Keeper
	ak auth.AccountKeeper
	pk params.Keeper
	vk Keeper

	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keyParams  *sdk.KVStoreKey
	tkeyParams *sdk.TransientStoreKey
	keyVM      *sdk.KVStoreKey

	pathBytes    []byte
	codeBytes    []byte
	addressBytes []byte
	valueBytes   []byte

	rawServer grpc.Server
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

func setupTestInput() testInput {
	input := testInput{
		cdc:        codec.New(),
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey("acc"),
		//keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyParams:  sdk.NewKVStoreKey("params"),
		tkeyParams: sdk.NewTransientStoreKey("transient_params"),
		keyVM:      sdk.NewKVStoreKey("vm"),
	}

	types.RegisterCodec(input.cdc)
	auth.RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyMain, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	mstore.MountStoreWithDB(input.keyVM, sdk.StoreTypeIAVL, db)
	err := mstore.LoadLatestVersion()
	if err != nil {
		panic(err)
	}

	input.pk = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams, params.DefaultCodespace)
	input.ak = auth.NewAccountKeeper(
		input.cdc,
		input.keyAccount,
		input.pk.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	config := vmConfig.DefaultVMConfig()

	var kpParams = keepalive.ClientParameters{
		Time:                time.Second,
		Timeout:             time.Second,
		PermitWithoutStream: true,
	}

	// no blocking, so we can launch part of tests even without vm
	clientConn, err := grpc.Dial(config.Address, grpc.WithInsecure(), grpc.WithKeepaliveParams(kpParams))
	if err != nil {
		panic(err)
	}

	listener, err := net.Listen("tcp", config.DataListen)
	if err != nil {
		panic(err)
	}

	input.vk = Keeper{
		cdc:      input.cdc,
		storeKey: input.keyVM,
		client:   vm_grpc.NewVMServiceClient(clientConn),
		listener: listener,
		config:   config,
	}

	input.vk.dsServer = NewDSServer(&input.vk)
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "wings-testnet-vm-keeper-test"}, false, log.NewNopLogger())

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

	return input
}

func LaunchVMMock() {

}
