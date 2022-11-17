package testutil

import (
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	client "github.com/cosmos/cosmos-sdk/x/quarantine/client/cli"
)

func (s *IntegrationTestSuite) TestTxOptInCmd() {
	addr0 := s.createAndFundAccount(0, 2000)
	s.Require().NoError(s.network.WaitForNextBlock(), "WaitForNextBlock")

	tests := []struct {
		name    string
		args    []string
		expErr  []string
		expCode int
	}{
		{
			name:   "empty addr",
			args:   []string{""},
			expErr: []string{"no to_name_or_address provided"},
		},
		{
			name:   "bad addr",
			args:   []string{"somethingelse"},
			expErr: []string{"somethingelse.info: key not found"},
		},
		{
			name:    "good addr",
			args:    []string{addr0},
			expCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.TxOptInCmd()
			cmdFuncName := "TxOptInCmd"
			args := append(tc.args, s.commonFlags...)
			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			out := outBW.String()
			s.T().Logf("Output:\n%s", out)
			s.assertErrorContents(err, tc.expErr, "%s error", cmdFuncName)
			for _, expErr := range tc.expErr {
				s.Assert().Contains(out, expErr, "%s output with error", cmdFuncName)
			}
			if len(tc.expErr) == 0 {
				var txResp sdk.TxResponse
				testFuncUn := func() {
					err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), &txResp)
				}
				if s.Assert().NotPanics(testFuncUn, "UnmarshalJSON output") {
					s.Assert().Equal(tc.expCode, int(txResp.Code), "%s response code", cmdFuncName)
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestTxOptOutCmd() {
	addr0 := s.createAndFundAccount(0, 2000)
	s.Require().NoError(s.network.WaitForNextBlock(), "WaitForNextBlock")

	tests := []struct {
		name    string
		args    []string
		expErr  []string
		expCode int
	}{
		{
			name:   "empty addr",
			args:   []string{""},
			expErr: []string{"no to_name_or_address provided"},
		},
		{
			name:   "bad addr",
			args:   []string{"somethingelse"},
			expErr: []string{"somethingelse.info: key not found"},
		},
		{
			name:    "good addr",
			args:    []string{addr0},
			expCode: 0,
		},
	}

	for _, tc := range tests {
		s.Run(tc.name, func() {
			cmd := client.TxOptOutCmd()
			cmdFuncName := "TxOptOutCmd"
			args := append(tc.args, s.commonFlags...)
			outBW, err := cli.ExecTestCLICmd(s.clientCtx, cmd, args)
			out := outBW.String()
			s.T().Logf("Output:\n%s", out)
			s.assertErrorContents(err, tc.expErr, "%s error", cmdFuncName)
			for _, expErr := range tc.expErr {
				s.Assert().Contains(out, expErr, "%s output with error", cmdFuncName)
			}
			if len(tc.expErr) == 0 {
				var txResp sdk.TxResponse
				testFuncUn := func() {
					err = s.clientCtx.Codec.UnmarshalJSON([]byte(out), &txResp)
				}
				if s.Assert().NotPanics(testFuncUn, "UnmarshalJSON output") {
					s.Assert().Equal(tc.expCode, int(txResp.Code), "%s response code", cmdFuncName)
				}
			}
		})
	}
}

// TODO[1046]: TxAcceptCmd()
// TODO[1046]: TxDeclineCmd()
// TODO[1046]: TxUpdateAutoResponsesCmd()
// TODO[1046]: SendAndAcceptQuarantinedFunds()
