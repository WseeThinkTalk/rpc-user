package logic

import (
	"context"

	"rpc-user/member/code"
	"rpc-user/member/internal/svc"
	"rpc-user/member/internal/types"
	"rpc-user/member/pb"

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

func (l *MemberOrderListLogic) MemberOrderList(in *pb.MemberOrderListRequest) (*pb.MemberOrderListResponse, error) {
	if in.UserId == 0 {
		return nil, code.UserIdEmpty
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
		return &pb.MemberOrderListResponse{IsEnd: true}, nil
	}

	items := make([]*pb.MemberOrderItem, 0, len(orders))
	for _, o := range orders {
		items = append(items, &pb.MemberOrderItem{
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

	cursor := orders[len(orders)-1].ID
	return &pb.MemberOrderListResponse{
		Items:  items,
		Cursor: cursor,
		IsEnd:  isEnd,
	}, nil
}
