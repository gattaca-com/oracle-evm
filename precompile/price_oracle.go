// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package precompile

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ava-labs/subnet-evm/vmerrs"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gattca/oracle-price-streamer/streamer"
)

type PriceFeedId common.Hash

var (
	_ StatefulPrecompileConfig = &PriceOracleConfig{}
	// Singleton StatefulPrecompiledContract for GetPriceing native assets by permissioned callers.
	PriceOraclePreCompile StatefulPrecompiledContract = CreateNativeGetPriceerPrecompile(PriceOracleAddress)

	// TODO perhaps put in a method to
	getPriceSignature = CalculateFunctionSelector("getPrice(uint256)")         // Hashed value of key (e.g. keccak256(btc/eth)) )
	setPriceSignature = CalculateFunctionSelector("setPrice(uint256,uint256)") // identitifer/key, new price price

	ErrCannotGetPrice = errors.New("non-enabled cannot GetPrice")

	GetPriceInputLen = common.HashLength
	SetPriceInputLen = common.HashLength + common.HashLength
)

var (
	AVAX_USD = PriceFeedId(common.BigToHash(big.NewInt(0)))
)

var SymbolToFeedId = map[string]PriceFeedId{
	"AVAX/USD": AVAX_USD,
}

func BytesToPriceFeedId(b []byte) PriceFeedId {
	return PriceFeedId(common.BytesToHash(b))
}

// PriceOracleConfig wraps [AllowListConfig] and uses it to implement the StatefulPrecompileConfig
// interface while adding in the contract deployer specific precompile address.
type PriceOracleConfig struct {
	BlockTimestamp  *big.Int         `json:"blockTimestamp"`
	AllowListAdmins []common.Address `json:"adminAddresses"`
}

// Address returns the address of the native GetPriceer contract.
func (c *PriceOracleConfig) Address() common.Address {
	return PriceOracleAddress
}

// Configure configures [state] with the desired admins based on [c].
func (c *PriceOracleConfig) Configure(state StateDB) {
	// DO configuration things in here
	// TODO
	if !state.Exist(c.Address()) {
		state.CreateAccount(c.Address())
	}

	state.SetState(c.Address(), common.BigToHash(big.NewInt(0)), common.BigToHash(big.NewInt(10000)))
}

// Contract returns the singleton stateful precompiled contract to be used for the native GetPriceer.
func (c *PriceOracleConfig) Contract() StatefulPrecompiledContract {
	return PriceOraclePreCompile
}

// Contract returns the singleton stateful precompiled contract to be used for the native GetPriceer.
func (c *PriceOracleConfig) Timestamp() *big.Int {
	return big.NewInt(0)
	// return big.NewInt(time.Now().AddDate(-1, 0, 0).UnixMilli())
	// return c.BlockTimestamp
}

// TODO PRETTIFY


func WritePriceToState(state StateDB, price *streamer.Price) {

	if !state.Exist(PriceOracleAddress) {
		state.CreateAccount(PriceOracleAddress)
	}

	if priceFeedId, ok := SymbolToFeedId[price.Symbol]; ok {
		state.SetState(PriceOracleAddress, common.Hash(priceFeedId), streamer.PriceToHash(price))
	}
}

/*
*
* Get Price Functionality Below
 */

// PackGetPriceInput packs [address] and [amount] into the appropriate arguments for GetPriceing operation.
func PackGetPriceInput(identifier *big.Int) ([]byte, error) {
	// function selector (4 bytes) + input(hash for address + hash for amount)
	fullLen := selectorLen + GetPriceInputLen
	input := make([]byte, fullLen)
	copy(input[:selectorLen], getPriceSignature)
	copy(input[selectorLen:selectorLen+common.HashLength], identifier.Bytes())
	return input, nil
}

// UnpackGetPriceInput attempts to unpack [input] into the arguments to the GetPrice precompile
// assumes that [input] does not include selector (omits first 4 bytes in PackGetPriceInput)
func UnpackGetPriceInput(input []byte) (*big.Int, error) {
	if len(input) != GetPriceInputLen {
		return nil, fmt.Errorf("invalid input length for GetPriceing: %d", len(input))
	}
	identifier := new(big.Int).SetBytes(input[:common.HashLength])
	return identifier, nil
}

// createGetPriceNativeCoin checks if the caller is permissioned for GetPriceing operation.
// The execution function parses the [input] into native coin amount and receiver address.
func getPrice(accessibleState PrecompileAccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = deductGas(suppliedGas, GetPriceGasCost); err != nil {
		return nil, 0, err
	}

	if readOnly {
		return nil, remainingGas, vmerrs.ErrWriteProtection
	}

	identifier, err := UnpackGetPriceInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	stateDB := accessibleState.GetStateDB()

	// if there is no address in the state, create one.
	if !stateDB.Exist(addr) {
		stateDB.CreateAccount(addr)
	}

	price := stateDB.GetState(addr, common.BigToHash(identifier))
	// Return an empty output and the remaining gas
	return price.Bytes(), remainingGas, nil
}

/*
*
* Set Price Functionality Below
 */

// PackGetPriceInput packs [address] and [amount] into the appropriate arguments for GetPriceing operation.
func PackSetPriceInput(identifier *big.Int, price *streamer.Price) ([]byte, error) {
	priceByte, err := streamer.MarshallPrice(price)
	if err != nil {
		return nil, err
	}

	priceLength := len(priceByte)

	if priceLength > 32 {
		return nil, fmt.Errorf("error Packing price. lenght of full price was larger than 32 bytes!. Total Lenth: %d", priceLength)
	}

	// function selector (4 bytes) + input(hash for address + hash for amount)
	fullLen := selectorLen + SetPriceInputLen
	input := make([]byte, fullLen)
	copy(input[:selectorLen], setPriceSignature)
	copy(input[selectorLen:selectorLen+common.HashLength], identifier.Bytes())
	copy(input[fullLen-common.HashLength:], priceByte)
	return input, nil
}

func UnpackSetPriceInput(input []byte) (*PriceFeedId, *streamer.Price, error) {
	if len(input) != SetPriceInputLen {
		return nil, nil, fmt.Errorf("invalid input length for SetPrice: %d", len(input))
	}
	identifier := BytesToPriceFeedId(input[:common.HashLength])
	price, err := streamer.UnmarshallPrice(input[common.HashLength : common.HashLength+common.HashLength])
	if err != nil {
		return nil, nil, err
	}
	return &identifier, price, nil
}

// SetPrice modifies the value set for that price, and sets it to a particular value
// The execution function parses the [input] into native coin amount and receiver address.
func setPrice(accessibleState PrecompileAccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = deductGas(suppliedGas, SetPriceGasCost); err != nil {
		return nil, 0, err
	}

	if readOnly {
		return nil, remainingGas, vmerrs.ErrWriteProtection
	}

	_, price, err := UnpackSetPriceInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	stateDB := accessibleState.GetStateDB()

	WritePriceToState(stateDB, price)

	// Return an empty output and the remaining gas
	return []byte{}, remainingGas, nil
}

// createNativeGetPriceerPrecompile returns a StatefulPrecompiledContract with R/W control of an allow list at [precompileAddr] and a native coin GetPriceer.
func CreateNativeGetPriceerPrecompile(precompileAddr common.Address) StatefulPrecompiledContract {
	GetPrice := newStatefulPrecompileFunction(getPriceSignature, getPrice)
	SetPrice := newStatefulPrecompileFunction(setPriceSignature, setPrice)

	// Construct the contract with no fallback function.
	contract := newStatefulPrecompileWithFunctionSelectors(nil, []*statefulPrecompileFunction{GetPrice, SetPrice})
	return contract
}
