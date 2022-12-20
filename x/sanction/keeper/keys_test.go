package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	"github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	"github.com/cosmos/cosmos-sdk/x/sanction/testutil"
)

func TestPrefixValues(t *testing.T) {
	prefixes := []struct {
		name     string
		prefix   []byte
		expected []byte
	}{
		{name: "ParamsPrefix", prefix: keeper.ParamsPrefix, expected: []byte{0x00}},
		{name: "SanctionedPrefix", prefix: keeper.SanctionedPrefix, expected: []byte{0x01}},
		{name: "TemporaryPrefix", prefix: keeper.TemporaryPrefix, expected: []byte{0x02}},
		{name: "ProposalIndexPrefix", prefix: keeper.ProposalIndexPrefix, expected: []byte{0x03}},
	}

	for _, p := range prefixes {
		t.Run(fmt.Sprintf("%s expected value", p.name), func(t *testing.T) {
			assert.Equal(t, p.prefix, p.expected, p.name)
		})
	}

	for i := 0; i < len(prefixes)-1; i++ {
		for j := i + 1; j < len(prefixes); j++ {
			t.Run(fmt.Sprintf("%s is different from %s", prefixes[i].name, prefixes[j].name), func(t *testing.T) {
				assert.NotEqual(t, prefixes[i].prefix, prefixes[j].prefix, "expected: %s, actual: %s", prefixes[i].name, prefixes[j].name)
			})
		}
	}
}

func TestConstValues(t *testing.T) {
	consts := []struct {
		name      string
		value     string
		exptected string
	}{
		{
			name:      "ParamNameImmediateSanctionMinDeposit",
			value:     keeper.ParamNameImmediateSanctionMinDeposit,
			exptected: "immediate_sanction_min_deposit",
		},
		{
			name:      "ParamNameImmediateUnsanctionMinDeposit",
			value:     keeper.ParamNameImmediateUnsanctionMinDeposit,
			exptected: "immediate_unsanction_min_deposit",
		},
	}

	for _, c := range consts {
		t.Run(fmt.Sprintf("%s expected value", c.name), func(t *testing.T) {
			assert.Equal(t, c.exptected, c.value)
		})
	}

	for i := 0; i < len(consts)-1; i++ {
		for j := i + 1; j < len(consts); j++ {
			t.Run(fmt.Sprintf("%s is different from %s", consts[i].name, consts[j].name), func(t *testing.T) {
				assert.NotEqual(t, consts[i].value, consts[j].value, "expected: %s, actual: %s", consts[i].name, consts[j].name)
			})
		}
	}
}

func TestConcatBz(t *testing.T) {
	type testCase struct {
		name     string
		bz1      []byte
		bz2      []byte
		expected []byte
	}
	copyTestCase := func(tc testCase) testCase {
		rv := testCase{
			name:     tc.name,
			bz1:      nil,
			bz2:      nil,
			expected: nil,
		}
		if tc.bz1 != nil {
			rv.bz1 = make([]byte, len(tc.bz1), cap(tc.bz1))
			copy(rv.bz1, tc.bz1)
		}
		if tc.bz2 != nil {
			rv.bz2 = make([]byte, len(tc.bz2), cap(tc.bz2))
			copy(rv.bz2, tc.bz2)
		}
		if tc.expected != nil {
			rv.expected = make([]byte, len(tc.expected), cap(tc.expected))
			copy(rv.expected, tc.expected)
		}
		return rv
	}

	tests := []testCase{
		{
			name:     "nil nil",
			bz1:      nil,
			bz2:      nil,
			expected: []byte{},
		},
		{
			name:     "nil empty",
			bz1:      nil,
			bz2:      []byte{},
			expected: []byte{},
		},
		{
			name:     "empty nil",
			bz1:      []byte{},
			bz2:      nil,
			expected: []byte{},
		},
		{
			name:     "empty empty",
			bz1:      []byte{},
			bz2:      []byte{},
			expected: []byte{},
		},
		{
			name:     "nil 1 byte",
			bz1:      nil,
			bz2:      []byte{'a'},
			expected: []byte{'a'},
		},
		{
			name:     "empty 1 byte",
			bz1:      []byte{},
			bz2:      []byte{'a'},
			expected: []byte{'a'},
		},
		{
			name:     "nil 4 bytes",
			bz1:      nil,
			bz2:      []byte("test"),
			expected: []byte("test"),
		},
		{
			name:     "empty 4 bytes",
			bz1:      []byte{},
			bz2:      []byte("test"),
			expected: []byte("test"),
		},
		{
			name:     "1 byte nil",
			bz1:      []byte{'a'},
			bz2:      nil,
			expected: []byte{'a'},
		},
		{
			name:     "1 byte empty",
			bz1:      []byte{'a'},
			bz2:      []byte{},
			expected: []byte{'a'},
		},
		{
			name:     "4 bytes nil",
			bz1:      []byte("test"),
			bz2:      nil,
			expected: []byte("test"),
		},
		{
			name:     "4 bytes empty",
			bz1:      []byte("test"),
			bz2:      []byte{},
			expected: []byte("test"),
		},
		{
			name:     "1 byte 1 byte",
			bz1:      []byte{'a'},
			bz2:      []byte{'b'},
			expected: []byte{'a', 'b'},
		},
		{
			name:     "1 byte 4 bytes",
			bz1:      []byte{'a'},
			bz2:      []byte("test"),
			expected: []byte("atest"),
		},
		{
			name:     "4 bytes 1 byte",
			bz1:      []byte("word"),
			bz2:      []byte{'x'},
			expected: []byte("wordx"),
		},
		{
			name:     "5 bytes 5 bytes",
			bz1:      []byte("hello"),
			bz2:      []byte("world"),
			expected: []byte("helloworld"),
		},
	}

	for _, tc_orig := range tests {
		passes := t.Run(tc_orig.name, func(t *testing.T) {
			tc := copyTestCase(tc_orig)
			var actual []byte
			testFunc := func() {
				actual = keeper.ConcatBz(tc.bz1, tc.bz2)
			}
			require.NotPanics(t, testFunc, "ConcatBz")
			assert.Equal(t, tc.expected, actual, "ConcatBz result")
			assert.Equal(t, len(tc.expected), len(actual), "ConcatBz result length")
			assert.Equal(t, cap(tc.expected), cap(actual), "ConcatBz result capacity")
			assert.Equal(t, len(actual), cap(actual), "ConcatBz result length and capacity")
			assert.Equal(t, tc_orig.bz1, tc.bz1, "input 1 before and after ConcatBz")
			assert.Equal(t, len(tc_orig.bz1), len(tc.bz1), "input 1 length before and after ConcatBz")
			assert.Equal(t, cap(tc_orig.bz1), cap(tc.bz1), "input 1 capacity before and after ConcatBz")
			assert.Equal(t, tc_orig.bz2, tc.bz2, "input 2 before and after ConcatBz")
			assert.Equal(t, len(tc_orig.bz2), len(tc.bz2), "input 2 length before and after ConcatBz")
			assert.Equal(t, cap(tc_orig.bz2), cap(tc.bz2), "input 2 capacity before and after ConcatBz")
			if cap(tc.bz1) > 0 {
				if len(tc.bz1) > 0 {
					if tc.bz1[0] == 'x' {
						tc.bz1[0] = 'y'
					} else {
						tc.bz1[0] = 'x'
					}
					assert.Equal(t, tc.expected, actual, "ConcatBz result after changing original bz1 input")
				}
				if len(tc.bz1) < cap(tc.bz1) {
					tc.bz1 = tc.bz1[:len(tc.bz1)+1]
					tc.bz1[len(tc.bz1)] = 'x'
					assert.Equal(t, tc.expected, actual, "ConcatBz result after extending original bz1 input")
				}
			}
			if cap(tc.bz2) > 0 {
				if len(tc.bz2) > 0 {
					if tc.bz2[0] == 'x' {
						tc.bz2[0] = 'y'
					} else {
						tc.bz2[0] = 'x'
					}
					assert.Equal(t, tc.expected, actual, "ConcatBz result after changing original bz2 input")
				}
				if len(tc.bz2) < cap(tc.bz2) {
					tc.bz2 = tc.bz2[:len(tc.bz2)+1]
					tc.bz2[len(tc.bz2)] = 'x'
					assert.Equal(t, tc.expected, actual, "ConcatBz result after extending original bz2 input")
				}
			}
		})
		if !passes {
			continue
		}

		if len(tc_orig.expected) > 0 {
			t.Run(tc_orig.name+" changing result", func(t *testing.T) {
				tc := copyTestCase(tc_orig)
				actual := keeper.ConcatBz(tc.bz1, tc.bz2)
				if len(actual) > 0 {
					if actual[0] == 'x' {
						actual[0] = 'y'
					} else {
						actual[0] = 'x'
					}
					assert.Equal(t, tc_orig.bz1, tc.bz1, "original bz1 after changing first result byte")
					assert.Equal(t, len(tc_orig.bz1), len(tc.bz1), "original bz1 length after changing first result byte")
					assert.Equal(t, cap(tc_orig.bz1), cap(tc.bz1), "original bz1 capacity after changing first result byte")
				}
				if len(actual) > 1 {
					if actual[len(actual)-1] == 'x' {
						actual[len(actual)-1] = 'y'
					} else {
						actual[len(actual)-1] = 'x'
					}
					assert.Equal(t, tc_orig.bz2, tc.bz2, "original bz2 after changing last result byte")
					assert.Equal(t, len(tc_orig.bz2), len(tc.bz2), "original bz2 length after changing last result byte")
					assert.Equal(t, cap(tc_orig.bz2), cap(tc.bz2), "original bz2 capacity after changing last result byte")
				}
			})
		}

		t.Run(tc_orig.name+" plus cap", func(t *testing.T) {
			tc := copyTestCase(tc_orig)
			plusCap := 5
			actual := keeper.ConcatBzPlusCap(tc.bz1, tc.bz2, plusCap)
			assert.Equal(t, tc.expected, actual, "concatBzPlusCap result")
			assert.Equal(t, len(tc.expected), len(actual), "concatBzPlusCap result length")
			assert.Equal(t, cap(tc.expected)+plusCap, cap(actual), "concatBzPlusCap result capacity")
			actual = actual[:len(actual)+1]
			actual[len(actual)-1] = 'x'
			assert.Equal(t, tc_orig.bz1, tc.bz1, "input 1 after extending result from concatBzPlusCap")
			assert.Equal(t, tc_orig.bz2, tc.bz2, "input 2 after extending result from concatBzPlusCap")
		})
	}
}

func TestParseLengthPrefixedBz(t *testing.T) {
	tests := []struct {
		name      string
		bz        []byte
		expAddr   []byte
		expSuffix []byte
		expPanic  string
	}{
		{
			name:     "nil",
			bz:       nil,
			expPanic: "expected key of length at least 1, got 0",
		},
		{
			name:     "empty",
			bz:       []byte{},
			expPanic: "expected key of length at least 1, got 0",
		},
		{
			name:      "only length byte of 0",
			bz:        []byte{0},
			expAddr:   []byte{},
			expSuffix: nil,
		},
		{
			name:     "only length byte of 1",
			bz:       []byte{1},
			expPanic: "expected key of length at least 2, got 1",
		},
		{
			name: "length byte 20 but one short",
			bz: []byte{20,
				'1', '2', '3', '4', '5', '6', '7', '8', '9', '0',
				'1', '2', '3', '4', '5', '6', '7', '8', '9',
			},
			expPanic: "expected key of length at least 21, got 20",
		},
		{
			name:      "length byte of 0 with extra",
			bz:        []byte{0, 'a', 'b', 'c'},
			expAddr:   []byte{},
			expSuffix: []byte("abc"),
		},
		{
			name:      "20 bytes no suffix",
			bz:        address.MustLengthPrefix([]byte("test_20_byte_addr___")),
			expAddr:   []byte("test_20_byte_addr___"),
			expSuffix: nil,
		},
		{
			name:      "20 bytes with suffix",
			bz:        append(address.MustLengthPrefix([]byte("test_20_byte_addr_2_")), []byte("something")...),
			expAddr:   sdk.AccAddress("test_20_byte_addr_2_"),
			expSuffix: []byte("something"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var addr, suffix []byte
			testFunc := func() {
				addr, suffix = keeper.ParseLengthPrefixedBz(tc.bz)
			}
			if len(tc.expPanic) > 0 {
				require.PanicsWithValue(t, tc.expPanic, testFunc, "ParseLengthPrefixedBz")
			} else {
				require.NotPanics(t, testFunc, "ParseLengthPrefixedBz")
				assert.Equal(t, tc.expAddr, addr, "ParseLengthPrefixedBz result addr")
				assert.Equal(t, tc.expSuffix, suffix, "ParseLengthPrefixedBz result suffix")
			}
		})
	}
}

func TestCreateParamKey(t *testing.T) {
	tests := []struct {
		name  string
		input string
		exp   []byte
	}{
		{
			name:  "control",
			input: "a word",
			exp:   append([]byte{keeper.ParamsPrefix[0]}, "a word"...),
		},
		{
			name:  "empty",
			input: "",
			exp:   keeper.ParamsPrefix,
		},
		{
			name:  "ParamNameImmediateSanctionMinDeposit",
			input: keeper.ParamNameImmediateSanctionMinDeposit,
			exp:   append([]byte{keeper.ParamsPrefix[0]}, keeper.ParamNameImmediateSanctionMinDeposit...),
		},
		{
			name:  "ParamNameImmediateUnsanctionMinDeposit",
			input: keeper.ParamNameImmediateUnsanctionMinDeposit,
			exp:   append([]byte{keeper.ParamsPrefix[0]}, keeper.ParamNameImmediateUnsanctionMinDeposit...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateParamKey(tc.input)
			}
			require.NotPanics(t, testFunc, "CreateParamKey")
			assert.Equal(t, tc.exp, actual, "CreateParamKey result")
		})
	}
}

func TestParseParamKey(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		exp   string
	}{
		{
			name:  "control",
			input: append([]byte{keeper.ParamsPrefix[0]}, "a word"...),
			exp:   "a word",
		},
		{
			name:  "empty",
			input: keeper.ParamsPrefix,
			exp:   "",
		},
		{
			name:  "ParamNameImmediateSanctionMinDeposit",
			input: keeper.CreateParamKey(keeper.ParamNameImmediateSanctionMinDeposit),
			exp:   keeper.ParamNameImmediateSanctionMinDeposit,
		},
		{
			name:  "ParamNameImmediateUnsanctionMinDeposit",
			input: keeper.CreateParamKey(keeper.ParamNameImmediateUnsanctionMinDeposit),
			exp:   keeper.ParamNameImmediateUnsanctionMinDeposit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual string
			testFunc := func() {
				actual = keeper.ParseParamKey(tc.input)
			}
			require.NotPanics(t, testFunc, "ParseParamKey")
			assert.Equal(t, tc.exp, actual, "ParseParamKey result")
		})
	}
}

func TestCreateSanctionedAddrKey(t *testing.T) {
	tests := []struct {
		name string
		addr sdk.AccAddress
		exp  []byte
	}{
		{
			name: "nil addr",
			addr: nil,
			exp:  []byte{keeper.SanctionedPrefix[0]},
		},
		{
			name: "4 byte address",
			addr: sdk.AccAddress("test"),
			exp:  append([]byte{keeper.SanctionedPrefix[0], 4}, "test"...),
		},
		{
			name: "20 byte address",
			addr: sdk.AccAddress("test_20_byte_address"),
			exp:  append([]byte{keeper.SanctionedPrefix[0], 20}, "test_20_byte_address"...),
		},
		{
			name: "32 byte address",
			addr: sdk.AccAddress("test_____32_____byte_____address"),
			exp:  append([]byte{keeper.SanctionedPrefix[0], 32}, "test_____32_____byte_____address"...),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = keeper.CreateSanctionedAddrKey(tc.addr)
			}
			require.NotPanics(t, testFunc, "CreateSanctionedAddrKey")
			assert.Equal(t, tc.exp, actual, "CreateSanctionedAddrKey result")
		})
	}
}

func TestParseSanctionedAddrKey(t *testing.T) {
	tests := []struct {
		name     string
		key      []byte
		exp      sdk.AccAddress
		expPanic string
	}{
		{
			name:     "nil",
			key:      nil,
			expPanic: "runtime error: slice bounds out of range [1:0]",
		},
		{
			name:     "empty",
			key:      []byte{},
			expPanic: "runtime error: slice bounds out of range [1:0]",
		},
		{
			name:     "just one byte",
			key:      []byte{'f'}, // doesn't matter what that byte is.
			expPanic: "expected key of length at least 1, got 0",
		},
		{
			name: "empty addr",
			key:  []byte{'g', 0},
			exp:  sdk.AccAddress{},
		},
		{
			name: "4 byte addr",
			key:  []byte{'P', 4, 't', 'e', 's', 't'},
			exp:  sdk.AccAddress("test"),
		},
		{
			name: "20 byte addr",
			key:  keeper.CreateSanctionedAddrKey(sdk.AccAddress("this_test_addr_is_20")),
			exp:  sdk.AccAddress("this_test_addr_is_20"),
		},
		{
			name: "32 byte addr",
			key:  keeper.CreateSanctionedAddrKey(sdk.AccAddress("this_test_addr_is_longer_with_32")),
			exp:  sdk.AccAddress("this_test_addr_is_longer_with_32"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.AccAddress
			testFunc := func() {
				actual = keeper.ParseSanctionedAddrKey(tc.key)
			}
			if len(tc.expPanic) > 0 {
				testutil.RequirePanicsWithMessage(t, tc.expPanic, testFunc, "ParseSanctionedAddrKey")
			} else {
				require.NotPanics(t, testFunc, "ParseSanctionedAddrKey")
				assert.Equal(t, tc.exp, actual, "ParseSanctionedAddrKey result")
			}
		})
	}
}

// TODO[1046]: CreateTemporaryAddrPrefix
// TODO[1046]: CreateTemporaryKey
// TODO[1046]: ParseTemporaryKey

func TestTempBValues(t *testing.T) {
	// If these were the same, it'd be bad.
	assert.NotEqual(t, keeper.TempSanctionB, keeper.TempUnsanctionB, "TempSanctionB vs TempUnsanctionB")
}

func TestIsTempSanctionBz(t *testing.T) {
	tests := []struct {
		name string
		bz   []byte
		exp  bool
	}{
		{name: "nil", bz: nil, exp: false},
		{name: "empty", bz: []byte{}, exp: false},
		{name: "TempSanctionB and 0", bz: []byte{keeper.TempSanctionB, 0}, exp: false},
		{name: "TempUnsanctionB and 0", bz: []byte{keeper.TempUnsanctionB, 0}, exp: false},
		{name: "0 and TempSanctionB", bz: []byte{0, keeper.TempSanctionB}, exp: false},
		{name: "0 and TempUnsanctionB", bz: []byte{0, keeper.TempUnsanctionB}, exp: false},
		{name: "the letter f", bz: []byte{'f'}, exp: false},
		{name: "TempSanctionB", bz: []byte{keeper.TempSanctionB}, exp: true},
		{name: "TempUnsanctionB", bz: []byte{keeper.TempUnsanctionB}, exp: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = keeper.IsTempSanctionBz(tc.bz)
			}
			require.NotPanics(t, testFunc, "IsTempSanctionBz")
			assert.Equal(t, tc.exp, actual, "IsTempSanctionBz result")
		})
	}
}

func TestIsTempUnsanctionBz(t *testing.T) {
	tests := []struct {
		name string
		bz   []byte
		exp  bool
	}{
		{name: "nil", bz: nil, exp: false},
		{name: "empty", bz: []byte{}, exp: false},
		{name: "TempSanctionB and 0", bz: []byte{keeper.TempSanctionB, 0}, exp: false},
		{name: "TempUnsanctionB and 0", bz: []byte{keeper.TempUnsanctionB, 0}, exp: false},
		{name: "0 and TempSanctionB", bz: []byte{0, keeper.TempSanctionB}, exp: false},
		{name: "0 and TempUnsanctionB", bz: []byte{0, keeper.TempUnsanctionB}, exp: false},
		{name: "the letter f", bz: []byte{'f'}, exp: false},
		{name: "TempSanctionB", bz: []byte{keeper.TempSanctionB}, exp: false},
		{name: "TempUnsanctionB", bz: []byte{keeper.TempUnsanctionB}, exp: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual bool
			testFunc := func() {
				actual = keeper.IsTempUnsanctionBz(tc.bz)
			}
			require.NotPanics(t, testFunc, "IsTempUnsanctionBz")
			assert.Equal(t, tc.exp, actual, "IsTempUnsanctionBz result")
		})
	}
}

func TestToTempStatus(t *testing.T) {
	tests := []struct {
		name string
		bz   []byte
		exp  sanction.TempStatus
	}{
		{name: "nil", bz: nil, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "empty", bz: []byte{}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "TempSanctionB and 0", bz: []byte{keeper.TempSanctionB, 0}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "TempUnsanctionB and 0", bz: []byte{keeper.TempUnsanctionB, 0}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "0 and TempSanctionB", bz: []byte{0, keeper.TempSanctionB}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "0 and TempUnsanctionB", bz: []byte{0, keeper.TempUnsanctionB}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "the letter f", bz: []byte{'f'}, exp: sanction.TEMP_STATUS_UNSPECIFIED},
		{name: "TempSanctionB", bz: []byte{keeper.TempSanctionB}, exp: sanction.TEMP_STATUS_SANCTIONED},
		{name: "TempUnsanctionB", bz: []byte{keeper.TempUnsanctionB}, exp: sanction.TEMP_STATUS_UNSANCTIONED},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sanction.TempStatus
			testFunc := func() {
				actual = keeper.ToTempStatus(tc.bz)
			}
			require.NotPanics(t, testFunc, "ToTempStatus")
			assert.Equal(t, tc.exp, actual, "ToTempStatus result")
		})
	}
}

// TODO[1046]: NewTempEvent
// TODO[1046]: CreateProposalTempIndexPrefix
// TODO[1046]: CreateProposalTempIndexKey
// TODO[1046]: ParseProposalTempIndexKey
