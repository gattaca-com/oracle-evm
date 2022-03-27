// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package precompile

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gattaca-com/OraclePriceStreamer/streamer"
	"github.com/gattaca-com/oracle-evm/vmerrs"
)

type PriceFeedId common.Hash

var (
	_ StatefulPrecompileConfig = &PriceOracleConfig{}
	// Singleton StatefulPrecompiledContract for GetPriceing native assets by permissioned callers.
	PriceOraclePreCompile StatefulPrecompiledContract = CreateNativeGetPriceerPrecompile(PriceOracleAddress)

	// TODO perhaps put in a method to
	getPriceSignature    = CalculateFunctionSelector("getPrice(uint256)")    // Hashed value of key (e.g. keccak256(btc/eth)) )
	getDecimalsSignature = CalculateFunctionSelector("getDecimals(uint256)") // Hashed value of key (e.g. keccak256(btc/eth)) )

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

func (p *PriceFeedId) Bytes() []byte {
	return common.Hash(*p).Bytes()
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

	sampleBtcAvaxVal := streamer.Price{
		Price:    10000,
		Slot:     12000,
		Symbol:   "AVAX/USD",
		Decimals: 8,
	}

	state.SetState(c.Address(), common.BigToHash(big.NewInt(0)), streamer.PriceToHash(&sampleBtcAvaxVal))
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

func WritePriceToState(state StateDB, price *streamer.Price) error {

	if !state.Exist(PriceOracleAddress) {
		state.CreateAccount(PriceOracleAddress)
	}

	if priceFeedId, ok := SymbolToFeedId[price.Symbol]; ok {
		state.SetState(PriceOracleAddress, common.Hash(priceFeedId), streamer.PriceToHash(price))
		return nil
	}

	return fmt.Errorf("Symbol id not currently supported to write. Key %s", price.Symbol)
}

/*
*
* Get Price Functionality Below
 */

// PackGetPriceInput packs [address] and [amount] into the appropriate arguments for GetPriceing operation.
func PackGetPriceInput(identifier *PriceFeedId) ([]byte, error) {
	// function selector (4 bytes) + input(hash for address + hash for amount)
	fullLen := selectorLen + GetPriceInputLen
	input := make([]byte, fullLen)
	copy(input[:selectorLen], getPriceSignature)
	copy(input[selectorLen:selectorLen+common.HashLength], identifier.Bytes())
	return input, nil
}

// UnpackGetPriceInput attempts to unpack [input] into the arguments to the GetPrice precompile
// assumes that [input] does not include selector (omits first 4 bytes in PackGetPriceInput)
func UnpackGetPriceInput(input []byte) (*PriceFeedId, error) {
	if len(input) != GetPriceInputLen {
		return nil, fmt.Errorf("invalid input length for GetPriceing: %d", len(input))
	}
	identifier := BytesToPriceFeedId(input[:common.HashLength])
	return &identifier, nil
}

func getPriceStruct(accessibleState PrecompileAccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (price *streamer.Price, remainingGas uint64, err error) {
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

	priceStructHash := stateDB.GetState(addr, common.Hash(*identifier))
	priceStruct, _ := streamer.UnmarshallPrice(priceStructHash.Bytes())

	return priceStruct, remainingGas, nil
}

func getPrice(accessibleState PrecompileAccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {

	priceStruct, remainingGas, err := getPriceStruct(accessibleState, caller, addr, input, suppliedGas, readOnly)
	if err != nil {
		return nil, remainingGas, err
	}

	price := big.NewInt(priceStruct.Price)
	// Return an empty output and the remaining gas
	return common.BigToHash(price).Bytes(), remainingGas, nil
}

func getDecimals(accessibleState PrecompileAccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {

	priceStruct, remainingGas, err := getPriceStruct(accessibleState, caller, addr, input, suppliedGas, readOnly)
	if err != nil {
		return nil, remainingGas, err
	}

	decimals := big.NewInt(int64(priceStruct.Decimals))
	// Return an empty output and the remaining gas
	return common.BigToHash(decimals).Bytes(), remainingGas, nil
}

// createNativeGetPriceerPrecompile returns a StatefulPrecompiledContract with R/W control of an allow list at [precompileAddr] and a native coin GetPriceer.
func CreateNativeGetPriceerPrecompile(precompileAddr common.Address) StatefulPrecompiledContract {
	GetPrice := newStatefulPrecompileFunction(getPriceSignature, getPrice)
	GetDecimals := newStatefulPrecompileFunction(getDecimalsSignature, getDecimals)

	// Construct the contract with no fallback function.
	contract := newStatefulPrecompileWithFunctionSelectors(nil, []*statefulPrecompileFunction{GetPrice, GetDecimals})
	return contract
}
