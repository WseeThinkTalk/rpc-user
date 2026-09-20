package memberlogic

import (
	"context"
	"fmt"
	"time"

	model "rpc-user/internal/model/member"
	"rpc-user/internal/svc"
	types "rpc-user/internal/types/member"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PayCallbackLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPayCallbackLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayCallbackLogic {
	return &PayCallbackLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PayCallbackLogic) PayCallback(in *user.PayCallbackRequest) (resp *user.PayCallbackResponse, err error) {
	resp = new(user.PayCallbackResponse)
	resp.Code = 200
	resp.Msg = "success"

	if in.OrderSn == "" {
		return nil, code.OrderSnEmpty
	}
	if in.TransactionId == "" {
		return nil, code.TransactionIdEmpty
	}

	order, err := l.svcCtx.MemberOrderModel.FindByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		l.Errorf("[PayCallback] FindByOrderSn err: %v, orderSn: %s", err, in.OrderSn)
		return nil, err
	}
	if order == nil {
		return nil, code.OrderNotFound
	}

	if order.Status == types.OrderStatusPaid {
		l.Infof("[PayCallback] order already paid, orderSn: %s, transactionId: %s", in.OrderSn, in.TransactionId)
		return resp, nil
	}

	now := time.Now()
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		orderModel := model.NewMemberOrderModel(tx)
		memberModel := model.NewMemberModel(tx)

		err := orderModel.UpdatePaySuccess(l.ctx, order.OrderSN, in.TransactionId, types.OrderStatusPaid)
		if err != nil {
			return fmt.Errorf("update order pay success err: %w", err)
		}

		currentMember, err := memberModel.FindByUserId(l.ctx, order.UserID)
		if err != nil {
			return fmt.Errorf("find member err: %w", err)
		}

		expireTime := now.Add(time.Duration(order.DurationDays) * 24 * time.Hour)
		if currentMember != nil && currentMember.Level == order.Level && currentMember.ExpireTime.After(now) {
			expireTime = currentMember.ExpireTime.Add(time.Duration(order.DurationDays) * 24 * time.Hour)
		}

		newMember := &model.Member{
			UserID:     order.UserID,
			Level:      order.Level,
			ExpireTime: expireTime,
			Status:     types.MemberStatusActive,
			CreateTime: now,
			UpdateTime: now,
		}
		return memberModel.UpsertMember(l.ctx, newMember)
	})
	if err != nil {
		l.Errorf("[PayCallback] Transaction err: %v, orderSn: %s", err, in.OrderSn)
		return nil, code.PaymentFailed
	}

	return resp, nil
}
