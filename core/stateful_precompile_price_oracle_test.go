package core

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ava-labs/subnet-evm/core/rawdb"
	"github.com/ava-labs/subnet-evm/core/state"
	"github.com/ava-labs/subnet-evm/precompile"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gattca/oracle-price-streamer/streamer"
)

type TestStateDb struct {
	mockstate map[common.Address]map[common.Hash]common.Hash
}

func NewTestStateDb() *TestStateDb {
	return &TestStateDb{mockstate: make(map[common.Address]map[common.Hash]common.Hash)}
}

func (tdb *TestStateDb) GetState(stateAddr common.Address, accessor common.Hash) common.Hash {
	if addressState, ok := tdb.mockstate[stateAddr]; ok {
		if state, ok := addressState[accessor]; ok {
			return state
		}
	}
	return common.Hash{}
}

func (tdb *TestStateDb) SetState(stateAddr common.Address, accessor common.Hash, value common.Hash) {
	if addressState, ok := tdb.mockstate[stateAddr]; ok {
		addressState[accessor] = value
	}
}

func (tdb *TestStateDb) SetCode(common.Address, []byte) {
	panic("Not implemented")
}

func (tdb *TestStateDb) SetNonce(common.Address, uint64) {
	panic("Not implemented")
}
func (tdb *TestStateDb) GetNonce(common.Address) uint64 {
	panic("Not implemented")
}

func (tdb *TestStateDb) GetBalance(common.Address) *big.Int {
	panic("Not implemented")
	// return big.NewInt(0)
}
func (tdb *TestStateDb) AddBalance(common.Address, *big.Int) {
	panic("Not implemented")
}
func (tdb *TestStateDb) SubBalance(common.Address, *big.Int) {
	panic("Not implemented")
}

func (tdb *TestStateDb) CreateAccount(account common.Address) {
	if _, ok := tdb.mockstate[account]; !ok {
		tdb.mockstate[account] = make(map[common.Hash]common.Hash)
	}
}

func (tdb *TestStateDb) Exist(account common.Address) bool {
	_, ok := tdb.mockstate[account]
	return ok
}

type TestPrecompileAccessibleState struct {
	tdb precompile.StateDB
}

func (t TestPrecompileAccessibleState) GetStateDB() precompile.StateDB {
	return t.tdb
}

func (t TestPrecompileAccessibleState) BlockTime() *big.Int {
	return big.NewInt(0)
}


func TestPriceOracleSetAndGetPrice(t *testing.T) {
	db := rawdb.NewMemoryDatabase()
	stateDb, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
	if err != nil {
		t.Fatal(err)
	}
	testPreCompileAccessibleState := TestPrecompileAccessibleState{stateDb}
	contract := precompile.CreateNativeGetPriceerPrecompile(precompile.PriceOracleAddress)

	sampleBtcAvaxVal := streamer.Price{
		Price: 10000,
		Slot: 12000,
		Symbol: "BTC/AVAX",
		Decimals: 8,
	}

	input, err := precompile.PackSetPriceInput(big.NewInt(0), &sampleBtcAvaxVal)

	if err != nil {
		t.Fatal(err)
	}

	_, _, err = contract.Run(testPreCompileAccessibleState, common.Address{}, precompile.PriceOracleAddress, input, 50000, false)

	if err != nil {
		t.Fatal(err)
	}

	// Now should be able to pull the price out
	input, err = precompile.PackGetPriceInput(big.NewInt(0))
	if err != nil {
		t.Fatal(err)
	}

	returnedVal, _, err := contract.Run(&testPreCompileAccessibleState, common.Address{}, precompile.PriceOracleAddress, input, 50000, false)
	if err != nil {
		t.Fatal(err)
	}

	returnedPrice, err := precompile.UnmarshallPrice(returnedVal)

	if err != nil {
		t.Fatal(err)
	}

	

	if reflect.DeepEqual(returnedPrice, &sampleBtcAvaxVal) {
		t.Errorf("Data was not stored or retreived correctly. Expected %+v. Returned %+v", returnedPrice, sampleBtcAvaxVal)
	}
}
