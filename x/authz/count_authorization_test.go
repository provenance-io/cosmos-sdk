package authz_test

import (
	"github.com/cosmos/cosmos-sdk/x/authz"
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
)

var (
	msgTypeURL                = "/cosmos.bank.v1beta1.MsgSend"
	zeroAllowedAuthorizations = int32(0)
	withAllowedAuthorizations = int32(2)
)

func TestCountAuthorization(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	authorization := authz.NewCountAuthorization(msgTypeURL, zeroAllowedAuthorizations)

	t.Log("verify authorization returns valid method name")
	require.Equal(t, authorization.MsgTypeURL(), "/cosmos.bank.v1beta1.MsgSend")
	require.NoError(t, authorization.ValidateBasic())

	t.Log("verify updated authorization returns nil")
	resp, err := authorization.Accept(ctx, nil)
	require.NoError(t, err)
	require.True(t, resp.Delete)
	require.Nil(t, resp.Updated)

	authorization = authz.NewCountAuthorization(msgTypeURL, withAllowedAuthorizations)
	require.Equal(t, authorization.MsgTypeURL(), "/cosmos.bank.v1beta1.MsgSend")
	require.NoError(t, authorization.ValidateBasic())

	t.Log("verify updated authorization returns correct count of allowed authorizations")
	resp, err = authorization.Accept(ctx, nil)
	require.NoError(t, err)
	require.False(t, resp.Delete)
	require.NotNil(t, resp.Updated)
	require.Equal(t, authorization.String(), resp.Updated.String())
}
