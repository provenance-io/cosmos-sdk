package testutil

func (s *IntegrationTestSuite) TestTxSanctionCmd() {
	// Cases:
	// no addresses given
	// one address good
	// one address bad
	// two addresses first bad
	// two addresses second bad
	// two addresses good
	// bad authority given
	// bad deposit
	//
	// Have each testcase have an expected proposal message.
	// If it's not nil, get the last proposal and check its contents.
	s.Fail("test not yet written")
}

func (s *IntegrationTestSuite) TestTxUnsanctionCmd() {
	// Cases:
	// no addresses given
	// one address good
	// one address bad
	// two addresses first bad
	// two addresses second bad
	// two addresses good
	// bad authority given
	// bad deposit
	//
	// Have each testcase have an expected proposal message.
	// If it's not nil, get the last proposal and check its contents.
	s.Fail("test not yet written")
}

func (s *IntegrationTestSuite) TestTxUpdateParamsCmd() {
	// Cases:
	// no args given
	// 1 arg given
	// 3 args given
	// coins coins
	// empty coins
	// coins empty
	// empty empty
	// bad coins
	// coins bad
	// bad authority given
	// bad deposit
	//
	// Have each testcase have an expected proposal message.
	// If it's not nil, get the last proposal and check its contents.
	s.Fail("test not yet written")
}
