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

type TestPrecompileAccessibleState struct {
	db precompile.StateDB
}

func (t TestPrecompileAccessibleState) GetStateDB() precompile.StateDB {
	return t.db
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

	returnedPrice, err := streamer.UnmarshallPrice(returnedVal)

	if err != nil {
		t.Fatal(err)
	}

	

	if reflect.DeepEqual(returnedPrice, &sampleBtcAvaxVal) {
		t.Errorf("Data was not stored or retreived correctly. Expected %+v. Returned %+v", returnedPrice, sampleBtcAvaxVal)
	}
}
