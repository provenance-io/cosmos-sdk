package quarantine

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrefixValues(t *testing.T) {
	assert.Len(t, OptInPrefix, 1, "OptInPrefix")
	assert.Len(t, AutoResponsePrefix, 1, "AutoResponsePrefix")
	assert.Len(t, RecordPrefix, 1, "RecordPrefix")
	assert.NotEqual(t, OptInPrefix[0], AutoResponsePrefix[0], "OptInPrefix vs AutoResponsePrefix")
	assert.NotEqual(t, OptInPrefix[0], RecordPrefix[0], "OptInPrefix vs RecordPrefix")
	assert.NotEqual(t, AutoResponsePrefix[0], RecordPrefix[0], "AutoResponsePrefix vs RecordPrefix")
}

func TestCreateOptInKey(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("coik test addr 0"),
		testAddr("coik test addr 1"),
	}
	badAddr := make(sdk.AccAddress, address.MaxAddrLen+1)
	for i := 0; i < len(badAddr); i++ {
		badAddr[i] = byte((i + 41) % 256)
	}
	makeExpected := func(pre []byte, addrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(addrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(len(addrBz)))
		rv = append(rv, addrBz...)
		return rv
	}
	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0",
			toAddr:   testAddrs[0],
			expected: makeExpected(OptInPrefix, testAddrs[0]),
		},
		{
			name:     "addr 0",
			toAddr:   testAddrs[1],
			expected: makeExpected(OptInPrefix, testAddrs[1]),
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: OptInPrefix,
		},
		{
			name:     "too long",
			toAddr:   badAddr,
			expected: nil,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = CreateOptInKey(tc.toAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateOptInKey") {
					assert.Equal(t, tc.expected, actual, "CreateOptInKey result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateOptInKey")
			}
		})
	}
}

func TestParseOptInKey(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("poik test addr 0"),
		testAddr("poik test addr 1"),
		testAddr("poik test addr 2"),
	}
	longAddr := make(sdk.AccAddress, 32)
	for i := 0; i < len(longAddr); i++ {
		longAddr[i] = byte((i + 65) % 256)
	}
	makeKey := func(pre []byte, addrLen int, addrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(addrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(addrLen))
		rv = append(rv, addrBz...)
		return rv
	}
	tests := []struct {
		name     string
		key      []byte
		expected sdk.AccAddress
		expPanic string
	}{
		{
			name:     "addr 0",
			key:      makeKey(OptInPrefix, len(testAddrs[0]), testAddrs[0]),
			expected: testAddrs[0],
		},
		{
			name:     "addr 1",
			key:      makeKey(OptInPrefix, len(testAddrs[1]), testAddrs[1]),
			expected: testAddrs[1],
		},
		{
			name:     "addr 2",
			key:      makeKey(OptInPrefix, len(testAddrs[2]), testAddrs[2]),
			expected: testAddrs[2],
		},
		{
			name:     "longer addr",
			key:      makeKey(OptInPrefix, len(longAddr), longAddr),
			expected: longAddr,
		},
		{
			name:     "too short",
			key:      makeKey(OptInPrefix, len(testAddrs[0])+1, testAddrs[0]),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", len(testAddrs[0])+1+2, len(testAddrs[0])+2),
		},
		{
			name:     "from CreateOptInKey addr 0",
			key:      CreateOptInKey(testAddrs[0]),
			expected: testAddrs[0],
		},
		{
			name:     "from CreateOptInKey addr 1",
			key:      CreateOptInKey(testAddrs[1]),
			expected: testAddrs[1],
		},
		{
			name:     "from CreateOptInKey addr 2",
			key:      CreateOptInKey(testAddrs[2]),
			expected: testAddrs[2],
		},
		{
			name:     "from CreateOptInKey longAddr",
			key:      CreateOptInKey(longAddr),
			expected: longAddr,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual sdk.AccAddress
			testFunc := func() {
				actual = ParseOptInKey(tc.key)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "ParseOptInKey") {
					assert.Equal(t, tc.expected, actual, "ParseOptInKey result")
				}
			} else {
				assert.PanicsWithValue(t, tc.expPanic, testFunc, "ParseOptInKey")
			}
		})
	}
}

func TestCreateAutoResponseToAddrPrefix(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("cartap test addr 0"),
		testAddr("cartap test addr 1"),
	}
	badAddr := make(sdk.AccAddress, address.MaxAddrLen+1)
	for i := 0; i < len(badAddr); i++ {
		badAddr[i] = byte((i + 45) % 256)
	}
	makeExpected := func(pre []byte, addrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(addrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(len(addrBz)))
		rv = append(rv, addrBz...)
		return rv
	}

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0",
			toAddr:   testAddrs[0],
			expected: makeExpected(AutoResponsePrefix, testAddrs[0]),
		},
		{
			name:     "addr 0",
			toAddr:   testAddrs[1],
			expected: makeExpected(AutoResponsePrefix, testAddrs[1]),
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: AutoResponsePrefix,
		},
		{
			name:     "too long",
			toAddr:   badAddr,
			expected: nil,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = CreateAutoResponseToAddrPrefix(tc.toAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateAutoResponseToAddrPrefix") {
					assert.Equal(t, tc.expected, actual, "CreateAutoResponseToAddrPrefix result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateAutoResponseToAddrPrefix")
			}
		})
	}
}

func TestCreateAutoResponseKey(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("cark test addr 0"),
		testAddr("cark test addr 1"),
	}
	badAddr := make(sdk.AccAddress, address.MaxAddrLen+1)
	for i := 0; i < len(badAddr); i++ {
		badAddr[i] = byte((i + 41) % 256)
	}
	longAddr := make(sdk.AccAddress, 32)
	for i := 0; i < len(longAddr); i++ {
		longAddr[i] = byte((i + 97) % 256)
	}
	makeExpected := func(pre []byte, toAddrBz, fromAddrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(toAddrBz)+1+len(fromAddrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(len(toAddrBz)))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(len(fromAddrBz)))
		rv = append(rv, fromAddrBz...)
		return rv
	}

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		fromAddr sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0 addr 1",
			toAddr:   testAddrs[0],
			fromAddr: testAddrs[1],
			expected: makeExpected(AutoResponsePrefix, testAddrs[0], testAddrs[1]),
		},
		{
			name:     "addr 1 long addr",
			toAddr:   testAddrs[1],
			fromAddr: longAddr,
			expected: makeExpected(AutoResponsePrefix, testAddrs[1], longAddr),
		},
		{
			name:     "long addr addr 0",
			toAddr:   longAddr,
			fromAddr: testAddrs[0],
			expected: makeExpected(AutoResponsePrefix, longAddr, testAddrs[0]),
		},
		{
			name:     "long addr long addr",
			toAddr:   longAddr,
			fromAddr: longAddr,
			expected: makeExpected(AutoResponsePrefix, longAddr, longAddr),
		},
		{
			name:     "bad toAddr",
			toAddr:   badAddr,
			fromAddr: testAddrs[0],
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
		{
			name:     "bad fromAddr",
			toAddr:   testAddrs[0],
			fromAddr: badAddr,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = CreateAutoResponseKey(tc.toAddr, tc.fromAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateAutoResponseKey") {
					assert.Equal(t, tc.expected, actual, "CreateAutoResponseKey result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateAutoResponseKey")
			}
		})
	}
}

func TestParseAutoResponseKey(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("park test addr 0"),
		testAddr("park test addr 1"),
	}
	longAddr := make(sdk.AccAddress, 32)
	for i := 0; i < len(longAddr); i++ {
		longAddr[i] = byte((i + 65) % 256)
	}
	makeKey := func(pre []byte, toAddrLen int, toAddrBz []byte, fromAddrLen int, fromAddrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(toAddrBz)+1+len(fromAddrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(toAddrLen))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(fromAddrLen))
		rv = append(rv, fromAddrBz...)
		return rv
	}

	tests := []struct {
		name        string
		key         []byte
		expToAddr   sdk.AccAddress
		expFromAddr sdk.AccAddress
		expPanic    string
	}{
		{
			name:        "addr 0 addr 1",
			key:         CreateAutoResponseKey(testAddrs[0], testAddrs[1]),
			expToAddr:   testAddrs[0],
			expFromAddr: testAddrs[1],
		},
		{
			name:        "addr 1 addr 0",
			key:         CreateAutoResponseKey(testAddrs[1], testAddrs[0]),
			expToAddr:   testAddrs[1],
			expFromAddr: testAddrs[0],
		},
		{
			name:        "long addr addr 1",
			key:         CreateAutoResponseKey(longAddr, testAddrs[1]),
			expToAddr:   longAddr,
			expFromAddr: testAddrs[1],
		},
		{
			name:        "addr 0 long addr",
			key:         CreateAutoResponseKey(testAddrs[0], longAddr),
			expToAddr:   testAddrs[0],
			expFromAddr: longAddr,
		},
		{
			name:     "bad toAddr len",
			key:      makeKey(AutoResponsePrefix, 200, testAddrs[0], 20, testAddrs[1]),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 202, 43),
		},
		{
			name:     "bad fromAddr len",
			key:      makeKey(AutoResponsePrefix, len(testAddrs[1]), testAddrs[1], len(testAddrs[0])+1, testAddrs[0]),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 44, 43),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actualToAddr, actualFromAddr sdk.AccAddress
			testFunc := func() {
				actualToAddr, actualFromAddr = ParseAutoResponseKey(tc.key)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "ParseAutoResponseKey") {
					assert.Equal(t, tc.expToAddr, actualToAddr, "ParseAutoResponseKey toAddr")
					assert.Equal(t, tc.expFromAddr, actualFromAddr, "ParseAutoResponseKey fromAddr")
				}
			} else {
				assert.PanicsWithValue(t, tc.expPanic, testFunc, "ParseAutoResponseKey")
			}
		})
	}
}

func TestCreateRecordToAddrPrefix(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("crtap test addr 0"),
		testAddr("crtap test addr 1"),
	}
	badAddr := make(sdk.AccAddress, address.MaxAddrLen+1)
	for i := 0; i < len(badAddr); i++ {
		badAddr[i] = byte((i + 45) % 256)
	}
	makeExpected := func(pre []byte, addrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(addrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(len(addrBz)))
		rv = append(rv, addrBz...)
		return rv
	}

	tests := []struct {
		name     string
		toAddr   sdk.AccAddress
		expected []byte
		expPanic string
	}{
		{
			name:     "addr 0",
			toAddr:   testAddrs[0],
			expected: makeExpected(RecordPrefix, testAddrs[0]),
		},
		{
			name:     "addr 0",
			toAddr:   testAddrs[1],
			expected: makeExpected(RecordPrefix, testAddrs[1]),
		},
		{
			name:     "nil",
			toAddr:   nil,
			expected: RecordPrefix,
		},
		{
			name:     "too long",
			toAddr:   badAddr,
			expected: nil,
			expPanic: fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = CreateRecordToAddrPrefix(tc.toAddr)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateRecordToAddrPrefix") {
					assert.Equal(t, tc.expected, actual, "CreateRecordToAddrPrefix result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateRecordToAddrPrefix")
			}
		})
	}
}

func TestCreateFromRecordId(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("cfri test addr 0"),
		testAddr("cfri test addr 0"),
		testAddr("cfri test addr 0"),
	}

	t.Run("panics if no addrs", func(t *testing.T) {
		assert.PanicsWithError(t, "at least one fromAddr is required: internal logic error",
			func() { createFromRecordId([]sdk.AccAddress{}) },
			"createFromRecordId([]sdk.AccAddress{})",
		)
	})

	t.Run("panics with nil addrs", func(t *testing.T) {
		assert.PanicsWithError(t, "at least one fromAddr is required: internal logic error",
			func() { createFromRecordId(nil) },
			"createFromRecordId(nil)",
		)
	})

	t.Run("single addrs are unchanged", func(t *testing.T) {
		for i, addr := range testAddrs {
			expected := make([]byte, len(addr))
			copy(expected, addr)

			actual := createFromRecordId([]sdk.AccAddress{addr})
			assert.Equal(t, expected, actual, "test address %d", i)
		}
	})

	t.Run("two addrs order does not matter", func(t *testing.T) {
		input1 := []sdk.AccAddress{testAddrs[0], testAddrs[1]}
		input2 := []sdk.AccAddress{testAddrs[1], testAddrs[2]}
		expected := createFromRecordId(input1)
		actual := createFromRecordId(input2)
		assert.Equal(t, expected, actual, "test addresses 0 and 1, vs 1 and 0")
	})

	t.Run("three addrs order does not matter", func(t *testing.T) {
		inputTestAddrsIndexes := [][]int{
			{0, 1, 2},
			{0, 2, 1},
			{1, 0, 2},
			{1, 2, 0},
			{2, 0, 1},
			{2, 1, 0},
		}
		inputs := make([][]sdk.AccAddress, len(inputTestAddrsIndexes))
		outputs := make([][]byte, len(inputTestAddrsIndexes))
		for i, taIndexes := range inputTestAddrsIndexes {
			inputs[i] = make([]sdk.AccAddress, len(taIndexes))
			for j, ind := range taIndexes {
				inputs[i][j] = testAddrs[ind]
			}
			outputs[i] = createFromRecordId(inputs[i])
		}
		for i := 0; i < len(outputs)-1; i++ {
			for j := i + 1; j < len(outputs); j++ {
				assert.Equal(t, outputs[i], outputs[j], "test addrs %v vs %v", inputTestAddrsIndexes[i], inputTestAddrsIndexes[j])
			}
		}
	})

	t.Run("two addrs different alone vs together", func(t *testing.T) {
		input1 := []sdk.AccAddress{testAddrs[1]}
		input2 := []sdk.AccAddress{testAddrs[2]}
		inputBoth := []sdk.AccAddress{testAddrs[1], testAddrs[2]}
		actual1 := createFromRecordId(input1)
		actual2 := createFromRecordId(input2)
		actualBoth := createFromRecordId(inputBoth)

		assert.NotEqual(t, actual1, actual2, "addr 1 vs addr 2")
		assert.NotEqual(t, actual1, actualBoth, "addr 1 vs both")
		assert.NotEqual(t, actual2, actualBoth, "addr 2 vs both")
		assert.NotContains(t, actualBoth, actual1, "both vs addr 1")
		assert.NotContains(t, actualBoth, actual2, "both vs addr 2")
	})
}

func TestCreateRecordKey(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("crk test addr 0"),
		testAddr("crk test addr 1"),
		testAddr("crk test addr 2"),
		testAddr("crk test addr 3"),
	}
	badAddr := make(sdk.AccAddress, address.MaxAddrLen+1)
	for i := 0; i < len(badAddr); i++ {
		badAddr[i] = byte((i + 41) % 256)
	}
	longAddr := make(sdk.AccAddress, 32)
	for i := 0; i < len(longAddr); i++ {
		longAddr[i] = byte((i + 97) % 256)
	}
	makeExpected := func(toAddrBz []byte, fromAddrs ...sdk.AccAddress) []byte {
		fromRecordId := createFromRecordId(fromAddrs)
		rv := make([]byte, 0, len(RecordPrefix)+1+len(toAddrBz)+1+len(fromRecordId))
		rv = append(rv, RecordPrefix...)
		rv = append(rv, byte(len(toAddrBz)))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(len(fromRecordId)))
		rv = append(rv, fromRecordId...)
		return rv
	}

	tests := []struct {
		name      string
		toAddr    sdk.AccAddress
		fromAddrs []sdk.AccAddress
		expected  []byte
		expPanic  string
	}{
		{
			name:      "addr 0 addr 1",
			toAddr:    testAddrs[0],
			fromAddrs: []sdk.AccAddress{testAddrs[1]},
			expected:  makeExpected(testAddrs[0], testAddrs[1]),
		},
		{
			name:      "addr 1 long addr",
			toAddr:    testAddrs[1],
			fromAddrs: []sdk.AccAddress{longAddr},
			expected:  makeExpected(testAddrs[1], longAddr),
		},
		{
			name:      "long addr addr 0",
			toAddr:    longAddr,
			fromAddrs: []sdk.AccAddress{testAddrs[0]},
			expected:  makeExpected(longAddr, testAddrs[0]),
		},
		{
			name:      "long addr long addr",
			toAddr:    longAddr,
			fromAddrs: []sdk.AccAddress{longAddr},
			expected:  makeExpected(longAddr, longAddr),
		},
		{
			name:      "to addr 3 from addrs 0 1 2 and long",
			toAddr:    testAddrs[3],
			fromAddrs: []sdk.AccAddress{testAddrs[0], testAddrs[1], testAddrs[2], longAddr},
			expected:  makeExpected(testAddrs[3], testAddrs[0], testAddrs[1], testAddrs[2], longAddr),
		},
		{
			name:      "to addr 2 from addrs 1 0 diff order",
			toAddr:    testAddrs[2],
			fromAddrs: []sdk.AccAddress{testAddrs[1], testAddrs[0]},
			expected:  makeExpected(testAddrs[2], testAddrs[0], testAddrs[1]),
		},
		{
			name:      "bad toAddr panics",
			toAddr:    badAddr,
			fromAddrs: []sdk.AccAddress{testAddrs[0]},
			expPanic:  fmt.Sprintf("address length should be max %d bytes, got %d: unknown address", address.MaxAddrLen, len(badAddr)),
		},
		{
			name:      "bad fromAddr ok",
			toAddr:    testAddrs[0],
			fromAddrs: []sdk.AccAddress{badAddr},
			expected:  makeExpected(testAddrs[0], badAddr),
		},
		{
			name:      "no fromAddrs panics",
			toAddr:    testAddrs[2],
			fromAddrs: []sdk.AccAddress{},
			expPanic:  "at least one fromAddr is required: internal logic error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actual []byte
			testFunc := func() {
				actual = CreateRecordKey(tc.toAddr, tc.fromAddrs...)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "CreateRecordKey") {
					assert.Equal(t, tc.expected, actual, "CreateRecordKey result")
				}
			} else {
				assert.PanicsWithError(t, tc.expPanic, testFunc, "CreateRecordKey")
			}
		})
	}
}

func TestParseRecordKey(t *testing.T) {
	testAddrs := []sdk.AccAddress{
		testAddr("prk test addr 0"),
		testAddr("prk test addr 1"),
		testAddr("prk test addr 2"),
	}
	longAddr := make(sdk.AccAddress, 32)
	for i := 0; i < len(longAddr); i++ {
		longAddr[i] = byte((i + 65) % 256)
	}
	makeKey := func(pre []byte, toAddrLen int, toAddrBz []byte, fromAddrLen int, fromAddrBz []byte) []byte {
		rv := make([]byte, 0, len(pre)+1+len(toAddrBz)+1+len(fromAddrBz))
		rv = append(rv, pre...)
		rv = append(rv, byte(toAddrLen))
		rv = append(rv, toAddrBz...)
		rv = append(rv, byte(fromAddrLen))
		rv = append(rv, fromAddrBz...)
		return rv
	}

	tests := []struct {
		name        string
		key         []byte
		expToAddr   sdk.AccAddress
		expFromAddr sdk.AccAddress
		expPanic    string
	}{
		{
			name:        "addr 0 addr 1",
			key:         CreateRecordKey(testAddrs[0], testAddrs[1]),
			expToAddr:   testAddrs[0],
			expFromAddr: testAddrs[1],
		},
		{
			name:        "addr 1 addr 0",
			key:         CreateRecordKey(testAddrs[1], testAddrs[0]),
			expToAddr:   testAddrs[1],
			expFromAddr: testAddrs[0],
		},
		{
			name:        "long addr addr 1",
			key:         CreateRecordKey(longAddr, testAddrs[1]),
			expToAddr:   longAddr,
			expFromAddr: testAddrs[1],
		},
		{
			name:        "addr 0 long addr",
			key:         CreateRecordKey(testAddrs[0], longAddr),
			expToAddr:   testAddrs[0],
			expFromAddr: longAddr,
		},
		{
			name:        "multiple from addrs",
			key:         CreateRecordKey(testAddrs[0], testAddrs[1], testAddrs[2]),
			expToAddr:   testAddrs[0],
			expFromAddr: createFromRecordId([]sdk.AccAddress{testAddrs[1], testAddrs[2]}),
		},
		{
			name:        "multiple from addrs diff order",
			key:         CreateRecordKey(testAddrs[0], testAddrs[2], testAddrs[1]),
			expToAddr:   testAddrs[0],
			expFromAddr: createFromRecordId([]sdk.AccAddress{testAddrs[1], testAddrs[2]}),
		},
		{
			name:     "bad toAddr len",
			key:      makeKey(RecordPrefix, 200, testAddrs[0], 20, testAddrs[1]),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 202, 43),
		},
		{
			name:     "bad fromAddr len",
			key:      makeKey(RecordPrefix, len(testAddrs[1]), testAddrs[1], len(testAddrs[0])+1, testAddrs[0]),
			expPanic: fmt.Sprintf("expected key of length at least %d, got %d", 44, 43),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var actualToAddr, actualFromAddr sdk.AccAddress
			testFunc := func() {
				actualToAddr, actualFromAddr = ParseRecordKey(tc.key)
			}
			if len(tc.expPanic) == 0 {
				if assert.NotPanics(t, testFunc, "ParseRecordKey") {
					assert.Equal(t, tc.expToAddr, actualToAddr, "ParseRecordKey toAddr")
					assert.Equal(t, tc.expFromAddr, actualFromAddr, "ParseRecordKey fromAddr")
				}
			} else {
				assert.PanicsWithValue(t, tc.expPanic, testFunc, "ParseRecordKey")
			}
		})
	}
}
