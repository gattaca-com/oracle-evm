// (c) 2019-2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package core

// import (
// 	"math/big"
// 	"strings"
// 	"testing"

// 	"github.com/gattaca-com/oracle-evm/core/rawdb"
// 	"github.com/gattaca-com/oracle-evm/core/state"
// 	"github.com/gattaca-com/oracle-evm/precompile"
// 	"github.com/gattaca-com/oracle-evm/vmerrs"
// 	"github.com/ethereum/go-ethereum/common"
// 	"github.com/ethereum/go-ethereum/common/math"
// 	"github.com/stretchr/testify/assert"
// )

// type mockAccessibleState struct {
// 	state *state.StateDB
// }

// func (m *mockAccessibleState) GetStateDB() precompile.StateDB { return m.state }

// This test is added within the core package so that it can import all of the required code
// without creating any import cycles
// func TestPriceOracleRun(t *testing.T) {
// 	type test struct {
// 		precompileAddr common.Address
// 		input          func() []byte
// 		suppliedGas    uint64
// 		readOnly       bool

// 		expectedRes []byte
// 		expectedErr string

// 		assertState func(t *testing.T, state *state.StateDB)
// 	}

// 	adminAddr := common.HexToAddress("0x8db97C7cEcE249c2b98bDC0226Cc4C2A57BF52FC")
// 	noRoleAddr := common.HexToAddress("0xF60C45c607D0f41687c94C314d300f483661E13a")

// 	for name, test := range map[string]test{
// 		"Set Price": {
// 			precompileAddr: precompile.PriceOracleAddress,
// 			input: func() []byte {
// 				input, err := precompile.PackSetPriceInput(big.NewInt(0), big.NewInt(1000))
// 				if err != nil {
// 					panic(err)
// 				}
// 				return input
// 			},
// 			suppliedGas: precompile.SetPriceGasCost,
// 			readOnly:    false,
// 			expectedRes: []byte{},
// 			assertState: func(t *testing.T, state *state.StateDB) {
// 				res := precompile.setPrice(state, adminAddr)
// 				assert.Equal(t, precompile.AllowListAdmin, res)

// 				res = precompile.GetContractDeployerAllowListStatus(state, noRoleAddr)
// 				assert.Equal(t, precompile.AllowListAdmin, res)
// 			},
// 		},
// 	} {
// 		t.Run(name, func(t *testing.T) {
// 			db := rawdb.NewMemoryDatabase()
// 			state, err := state.New(common.Hash{}, state.NewDatabase(db), nil)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			ret, remainingGas, err := precompile.ContractNativeMinterPrecompile.Run(&mockAccessibleState{state: state}, test.caller, test.precompileAddr, test.input(), test.suppliedGas, test.readOnly)
// 			if len(test.expectedErr) != 0 {
// 				if err == nil {
// 					assert.Failf(t, "run expectedly passed without error", "expected error %q", test.expectedErr)
// 				} else {
// 					assert.True(t, strings.Contains(err.Error(), test.expectedErr), "expected error (%s) to contain substring (%s)", err, test.expectedErr)
// 				}
// 				return
// 			}

// 			if err != nil {
// 				t.Fatal(err)
// 			}

// 			assert.Equal(t, uint64(0), remainingGas)
// 			assert.Equal(t, test.expectedRes, ret)

// 			test.assertState(t, state)
// 		})
// 	}
// }
