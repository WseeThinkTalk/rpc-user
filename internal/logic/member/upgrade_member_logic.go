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
	resp.Code = 200
	resp.Msg = "success"

	if in.UserId == 0 {
		return nil, code.MemberUserIdEmpty
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
		l.Errorf("[UpgradeMember] Transaction err: %v userId: %d", err, in.UserId)
		return nil, code.PaymentFailed
	}

	return resp, nil
}
