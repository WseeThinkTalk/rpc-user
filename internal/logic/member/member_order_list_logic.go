package memberlogic

import (
	"context"

	"rpc-user/internal/svc"
	types "rpc-user/internal/types/member"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type MemberOrderListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMemberOrderListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MemberOrderListLogic {
	return &MemberOrderListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *MemberOrderListLogic) MemberOrderList(in *user.MemberOrderListRequest) (resp *user.MemberOrderListResponse, err error) {
	resp = new(user.MemberOrderListResponse)
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = new(user.MemberOrderListData)
	resp.Data.Items = make([]*user.MemberOrderItem, 0)

	if in.UserId == 0 {
		return nil, code.MemberUserIdEmpty
	}
	if in.PageSize == 0 {
		in.PageSize = types.DefaultPageSize
	}

	orders, err := l.svcCtx.MemberOrderModel.FindByUserId(l.ctx, in.UserId, in.Cursor, in.PageSize+1)
	if err != nil {
		l.Errorf("[MemberOrderList] FindByUserId err: %v userId: %d", err, in.UserId)
		return nil, err
	}

	var isEnd bool
	if len(orders) > int(in.PageSize) {
		orders = orders[:in.PageSize]
	} else {
		isEnd = true
	}
	if len(orders) == 0 {
		resp.Data.IsEnd = true
		return resp, nil
	}

	for _, o := range orders {
		resp.Data.Items = append(resp.Data.Items, &user.MemberOrderItem{
			Id:           o.ID,
			UserId:       o.UserID,
			Level:        o.Level,
			DurationDays: o.DurationDays,
			Amount:       o.Amount,
			PayChannel:   o.PayChannel,
			Status:       o.Status,
			CreateTime:   o.CreateTime.Unix(),
		})
	}

	nextCursor := in.Cursor
	if len(orders) > 0 {
		nextCursor = orders[len(orders)-1].ID
	}

	resp.Data.Cursor = nextCursor
	resp.Data.IsEnd = isEnd

	return resp, nil
}
