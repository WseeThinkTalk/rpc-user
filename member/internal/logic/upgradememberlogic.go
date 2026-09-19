package logic

import (
	"context"
	"fmt"
	"time"

	"rpc-user/member/code"
	"rpc-user/member/internal/model"
	"rpc-user/member/internal/svc"
	"rpc-user/member/internal/types"
	"rpc-user/member/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpgradeMemberLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpgradeMemberLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpgradeMemberLogic {
	return &UpgradeMemberLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *UpgradeMemberLogic) UpgradeMember(in *pb.UpgradeMemberRequest) (*pb.UpgradeMemberResponse, error) {
	if in.UserId == 0 {
		return nil, code.UserIdEmpty
	}
	if in.Level < types.MemberLevelGold || in.Level > types.MemberLevelDiamond {
		return nil, code.LevelInvalid
	}
	if in.TransactionId == "" {
		return nil, code.TransactionIdEmpty
	}

	existing, err := l.svcCtx.MemberOrderModel.FindByTransactionId(l.ctx, in.TransactionId)
	if err != nil {
		l.Errorf("[UpgradeMember] FindByTransactionId err: %v txId: %s", err, in.TransactionId)
		return nil, err
	}
	if existing != nil {
		return nil, code.DuplicateTransaction
	}

	now := time.Now()
	expireTime := now.Add(time.Duration(in.DurationDays) * 24 * time.Hour)

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		order := &model.MemberOrder{
			UserID:        in.UserId,
			Level:         in.Level,
			DurationDays:  in.DurationDays,
			Amount:        in.Amount,
			PayChannel:    in.PayChannel,
			TransactionID: in.TransactionId,
			Status:        types.OrderStatusPaid,
			CreateTime:    now,
			UpdateTime:    now,
		}
		if err := model.NewMemberOrderModel(tx).Insert(l.ctx, order); err != nil {
			return err
		}

		member := &model.Member{
			UserID:     in.UserId,
			Level:      in.Level,
			ExpireTime: expireTime,
			Status:     types.MemberStatusActive,
			CreateTime: now,
			UpdateTime: now,
		}
		return model.NewMemberModel(tx).UpsertMember(l.ctx, member)
	})
	if err != nil {
		l.Errorf("[UpgradeMember] transaction err: %v req: %+v", err, in)
		return nil, err
	}

	key := memberInfoKey(in.UserId)
	if _, err := l.svcCtx.BizRedis.DelCtx(l.ctx, key); err != nil {
		l.Errorf("[UpgradeMember] redis del err: %v key: %s", err, key)
	}

	return &pb.UpgradeMemberResponse{}, nil
}

func memberInfoKey(userId int64) string {
	return fmt.Sprintf("biz#member#info#%d", userId)
}
