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

	if in.OrderSn == "" {
		resp.Code = int64(code.OrderSnEmpty.Code())
		resp.Msg = code.OrderSnEmpty.Message()
		return resp, nil
	}
	if in.TransactionId == "" {
		resp.Code = int64(code.TransactionIdEmpty.Code())
		resp.Msg = code.TransactionIdEmpty.Message()
		return resp, nil
	}

	order, err := l.svcCtx.MemberOrderModel.FindByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}
	if order == nil {
		resp.Code = int64(code.OrderNotFound.Code())
		resp.Msg = code.OrderNotFound.Message()
		return resp, nil
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
		resp.Code = int64(code.PaymentFailed.Code())
		resp.Msg = code.PaymentFailed.Message()
		return resp, nil
	}

	return resp, nil
}
