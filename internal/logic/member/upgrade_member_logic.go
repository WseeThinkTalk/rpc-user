package memberlogic

import (
	"context"
	"time"

	model "rpc-user/internal/model/member"
	"rpc-user/internal/svc"
	types "rpc-user/internal/types/member"
	"rpc-user/pkg/code"
	"rpc-user/user"

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

func (l *UpgradeMemberLogic) UpgradeMember(in *user.UpgradeMemberRequest) (resp *user.UpgradeMemberResponse, err error) {
	resp = new(user.UpgradeMemberResponse)

	if in.UserId == 0 {
		resp.Code = int64(code.MemberUserIdEmpty.Code())
		resp.Msg = code.MemberUserIdEmpty.Message()
		return resp, nil
	}
	if in.Level < types.MemberLevelGold || in.Level > types.MemberLevelDiamond {
		resp.Code = int64(code.LevelInvalid.Code())
		resp.Msg = code.LevelInvalid.Message()
		return resp, nil
	}
	if in.TransactionId == "" {
		resp.Code = int64(code.TransactionIdEmpty.Code())
		resp.Msg = code.TransactionIdEmpty.Message()
		return resp, nil
	}

	existing, err := l.svcCtx.MemberOrderModel.FindByTransactionId(l.ctx, in.TransactionId)
	if err != nil {
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}
	if existing != nil {
		resp.Code = int64(code.DuplicateTransaction.Code())
		resp.Msg = code.DuplicateTransaction.Message()
		return resp, nil
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
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		current, err := l.svcCtx.MemberModel.FindByUserId(l.ctx, in.UserId)
		if err != nil {
			return err
		}

		if current == nil {
			return tx.Create(&model.Member{
				UserID:     in.UserId,
				Level:      in.Level,
				ExpireTime: expireTime,
				Status:     types.MemberStatusActive,
				CreateTime: now,
				UpdateTime: now,
			}).Error
		}

		if current.ExpireTime.After(now) {
			expireTime = current.ExpireTime.Add(time.Duration(in.DurationDays) * 24 * time.Hour)
		}
		return tx.Model(current).Updates(map[string]interface{}{
			"level":       in.Level,
			"expire_time": expireTime,
			"status":      types.MemberStatusActive,
			"update_time": now,
		}).Error
	})
	if err != nil {
		resp.Code = int64(code.PaymentFailed.Code())
		resp.Msg = code.PaymentFailed.Message()
		return resp, nil
	}

	return resp, nil
}
