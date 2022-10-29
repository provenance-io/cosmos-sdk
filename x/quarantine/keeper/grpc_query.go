package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ quarantine.QueryServer = Keeper{}

func (k Keeper) QuarantinedFunds(goCtx context.Context, req *quarantine.QueryQuarantinedFundsRequest) (*quarantine.QueryQuarantinedFundsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.FromAddress) > 0 && len(req.ToAddress) == 0 {
		return nil, status.Error(codes.InvalidArgument, "to address cannot be empty when from address is not")
	}

	var toAddr, fromAddr sdk.AccAddress
	var err error
	if len(req.ToAddress) > 0 {
		toAddr, err = sdk.AccAddressFromBech32(req.ToAddress)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid to address: %s", err.Error())
		}
	}
	if len(req.ToAddress) > 0 {
		fromAddr, err = sdk.AccAddressFromBech32(req.FromAddress)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid from address: %s", err.Error())
		}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	resp := &quarantine.QueryQuarantinedFundsResponse{}

	if len(toAddr) > 0 && len(fromAddr) > 0 {
		qr := k.GetQuarantineRecord(ctx, toAddr, fromAddr)
		qf := qr.AsQuarantinedFunds(toAddr, fromAddr)
		resp.QuarantinedFunds = append(resp.QuarantinedFunds, qf)
	} else {
		store := k.getQuarantineRecordPrefixStore(ctx, toAddr)
		resp.Pagination, err = query.FilteredPaginate(
			store, req.Pagination,
			func(key, value []byte, accumulate bool) (bool, error) {
				var qr quarantine.QuarantineRecord
				qr, err = k.bzToQuarantineRecord(value)
				if err != nil {
					return false, err
				}
				if qr.Declined {
					return false, nil
				}
				if accumulate {
					kToAddr, kFromAddr := quarantine.ParseRecordKey(key)
					qf := qr.AsQuarantinedFunds(kToAddr, kFromAddr)
					resp.QuarantinedFunds = append(resp.QuarantinedFunds, qf)
				}
				return true, nil
			},
		)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}

func (k Keeper) IsQuarantined(goCtx context.Context, req *quarantine.QueryIsQuarantinedRequest) (*quarantine.QueryIsQuarantinedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.ToAddress) == 0 {
		return nil, status.Error(codes.InvalidArgument, "to address cannot be empty")
	}

	toAddr, err := sdk.AccAddressFromBech32(req.ToAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid to address: %s", err.Error())
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &quarantine.QueryIsQuarantinedResponse{
		IsQuarantined: k.IsQuarantinedAddr(ctx, toAddr),
	}

	return resp, nil
}

func (k Keeper) AutoResponses(goCtx context.Context, req *quarantine.QueryAutoResponsesRequest) (*quarantine.QueryAutoResponsesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.ToAddress) == 0 {
		return nil, status.Error(codes.InvalidArgument, "to address cannot be empty")
	}

	toAddr, err := sdk.AccAddressFromBech32(req.ToAddress)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid to address: %s", err.Error())
	}

	var fromAddr sdk.AccAddress
	if len(req.FromAddress) > 0 {
		fromAddr, err = sdk.AccAddressFromBech32(req.FromAddress)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid from address: %s", err.Error())
		}
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	resp := &quarantine.QueryAutoResponsesResponse{}

	if len(fromAddr) > 0 {
		qar := k.GetAutoResponse(ctx, toAddr, fromAddr)
		r := quarantine.NewAutoResponseEntry(toAddr, fromAddr, qar)
		resp.Results = append(resp.Results, r)
	} else {
		store := k.getAutoResponsesPrefixStore(ctx, toAddr)
		resp.Pagination, err = query.Paginate(
			store, req.Pagination,
			func(key, value []byte) error {
				kToAddr, kFromAddr := quarantine.ParseAutoResponseKey(key)
				qar := quarantine.ToAutoResponse(value)
				r := quarantine.NewAutoResponseEntry(kToAddr, kFromAddr, qar)
				resp.Results = append(resp.Results, r)
				return nil
			},
		)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return resp, nil
}
