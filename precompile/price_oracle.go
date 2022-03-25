// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package precompile

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ava-labs/subnet-evm/vmerrs"
	"github.com/ethereum/go-ethereum/common"
)

var (
	_ StatefulPrecompileConfig = &PriceOracleConfig{}
	// Singleton StatefulPrecompiledContract for GetPriceing native assets by permissioned callers.
	PriceOraclePreCompile StatefulPrecompiledContract = CreateNativeGetPriceerPrecompile(PriceOracleAddress)

	// TODO perhaps put in a method to
	getPriceSignature = CalculateFunctionSelector("getPrice(uint256)") // Hashed value of key (e.g. keccak256(btc/eth)) )
	setPriceSignature = CalculateFunctionSelector("setPrice(uint256,uint256)") // identitifer/key, new price price

	ErrCannotGetPrice = errors.New("non-enabled cannot GetPrice")

	GetPriceInputLen = common.HashLength
	SetPriceInputLen = common.HashLength + common.HashLength
)

// PriceOracleConfig wraps [AllowListConfig] and uses it to implement the StatefulPrecompileConfig
// interface while adding in the contract deployer specific precompile address.
type PriceOracleConfig struct {
	AllowListConfig
}

// Address returns the address of the native GetPriceer contract.
func (c *PriceOracleConfig) Address() common.Address {
	return PriceOracleAddress
}

// Configure configures [state] with the desired admins based on [c].
func (c *PriceOracleConfig) Configure(state StateDB) {
	c.AllowListConfig.Configure(state, PriceOracleAddress)
}

// Contract returns the singleton stateful precompiled contract to be used for the native GetPriceer.
func (c *PriceOracleConfig) Contract() StatefulPrecompiledContract {
	return PriceOraclePreCompile
}

// GetPriceOracleStatus returns the role of [address] for the GetPriceer list.
func GetPriceOracleStatus(stateDB StateDB, address common.Address) AllowListRole {
	return getAllowListStatus(stateDB, PriceOracleAddress, address)
}

// SetPriceOracleStatus sets the permissions of [address] to [role] for the
// GetPriceer list. assumes [role] has already been verified as valid.
func SetPriceOracleStatus(stateDB StateDB, address common.Address, role AllowListRole) {
	setAllowListRole(stateDB, PriceOracleAddress, address, role)
}

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
func createGetPrice(accessibleState PrecompileAccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = deductGas(suppliedGas, GetPriceGasCost); err != nil {
		return nil, 0, err
	}

	// if readOnly {
	// 	return nil, remainingGas, vmerrs.ErrWriteProtection
	// }

	identifier, err := UnpackGetPriceInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	stateDB := accessibleState.GetStateDB()
	// Verify that the caller is in the allow list and therefore has the right to modify it
	// callerStatus := getAllowListStatus(stateDB, PriceOracleAddress, caller)
	// if !callerStatus.IsEnabled() {
	// 	return nil, remainingGas, fmt.Errorf("%w: %s", ErrCannotGetPrice, caller)
	// }

	// if there is no address in the state, create one.
	if !stateDB.Exist(addr) {
		stateDB.CreateAccount(addr)
	}

	price := stateDB.GetState(addr, common.BigToHash(identifier))
	// Return an empty output and the remaining gas
	return price.Bytes(), remainingGas, nil
}


// PackGetPriceInput packs [address] and [amount] into the appropriate arguments for GetPriceing operation.
func PackSetPriceInput(identifier []byte, amount *big.Int) ([]byte, error) {
	// function selector (4 bytes) + input(hash for address + hash for amount)
	fullLen := selectorLen + SetPriceInputLen
	input := make([]byte, fullLen)
	copy(input[:selectorLen], setPriceSignature)
	copy(input[selectorLen:selectorLen+common.HashLength], identifier)
	amount.FillBytes(input[fullLen-common.HashLength: fullLen])
	return input, nil
}

func UnpackSetPriceInput(input []byte) (common.Hash, *big.Int, error) {
	if len(input) != SetPriceInputLen {
		return common.Hash{}, nil, fmt.Errorf("invalid input length for SetPrice: %d", len(input))
	}
	identifier := common.BytesToHash(input[:common.HashLength])
	assetAmount := new(big.Int).SetBytes(input[common.HashLength : common.HashLength+common.HashLength])
	return identifier, assetAmount, nil
}

// SetPrice modifies the value set for that price, and sets it to a particular value
// The execution function parses the [input] into native coin amount and receiver address.
func createSetPrice(accessibleState PrecompileAccessibleState, caller common.Address, addr common.Address, input []byte, suppliedGas uint64, readOnly bool) (ret []byte, remainingGas uint64, err error) {
	if remainingGas, err = deductGas(suppliedGas, SetPriceGasCost); err != nil {
		return nil, 0, err
	}

	if readOnly {
		return nil, remainingGas, vmerrs.ErrWriteProtection
	}

	to, amount, err := UnpackSetPriceInput(input)
	if err != nil {
		return nil, remainingGas, err
	}

	stateDB := accessibleState.GetStateDB()
	// Verify that the caller is in the allow list and therefore has the right to modify it
	// callerStatus := getAllowListStatus(stateDB, PriceOracleAddress, caller)
	// if !callerStatus.IsEnabled() {
	// 	return nil, remainingGas, fmt.Errorf("%w: %s", ErrCannotGetPrice, caller)
	// }

	// if there is no address in the state, create one.
	if !stateDB.Exist(addr) {
		stateDB.CreateAccount(addr)
	}

	stateDB.SetState(addr, to, common.BigToHash(amount))
	// Return an empty output and the remaining gas
	return []byte{}, remainingGas, nil
}

// createNativeGetPriceerPrecompile returns a StatefulPrecompiledContract with R/W control of an allow list at [precompileAddr] and a native coin GetPriceer.
func CreateNativeGetPriceerPrecompile(precompileAddr common.Address) StatefulPrecompiledContract {
	setAdmin := newStatefulPrecompileFunction(setAdminSignature, createAllowListRoleSetter(precompileAddr, AllowListAdmin))
	setEnabled := newStatefulPrecompileFunction(setEnabledSignature, createAllowListRoleSetter(precompileAddr, AllowListEnabled))
	setNone := newStatefulPrecompileFunction(setNoneSignature, createAllowListRoleSetter(precompileAddr, AllowListNoRole))
	read := newStatefulPrecompileFunction(readAllowListSignature, createReadAllowList(precompileAddr))

	GetPrice := newStatefulPrecompileFunction(getPriceSignature, createGetPrice)
	SetPrice := newStatefulPrecompileFunction(setPriceSignature, createSetPrice)

	// Construct the contract with no fallback function.
	contract := newStatefulPrecompileWithFunctionSelectors(nil, []*statefulPrecompileFunction{setAdmin, setEnabled, setNone, read, GetPrice, SetPrice})
	return contract
}