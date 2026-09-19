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

func (l *PayCallbackLogic) PayCallback(in *pb.PayCallbackRequest) (*pb.PayCallbackResponse, error) {
	if in.OrderSn == "" {
		return nil, code.OrderSnEmpty
	}
	if in.TransactionId == "" {
		return nil, code.TransactionIdEmpty
	}

	// 1. 查询订单
	order, err := l.svcCtx.MemberOrderModel.FindByOrderSn(l.ctx, in.OrderSn)
	if err != nil {
		l.Errorf("[PayCallback] FindByOrderSn err: %v, orderSn: %s", err, in.OrderSn)
		return nil, err
	}
	if order == nil {
		return nil, code.OrderNotFound
	}

	// 2. 幂等性校验，如果已支付则直接返回成功
	if order.Status == types.OrderStatusPaid {
		l.Infof("[PayCallback] order already paid, orderSn: %s, transactionId: %s", in.OrderSn, in.TransactionId)
		return &pb.PayCallbackResponse{}, nil
	}

	// 3. 事务处理：更新订单状态并更新/开通会员
	now := time.Now()
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// a. 更新订单为已支付状态并填写第三方支付流水号
		orderModel := model.NewMemberOrderModel(tx)
		if err := orderModel.UpdatePaySuccess(l.ctx, in.OrderSn, in.TransactionId, types.OrderStatusPaid); err != nil {
			return err
		}

		// b. 查询当前用户的会员记录
		memberModel := model.NewMemberModel(tx)
		member, err := memberModel.FindByUserId(l.ctx, order.UserID)
		if err != nil {
			return err
		}

		// c. 计算会员过期时间
		var expireTime time.Time
		if member != nil && member.Status == types.MemberStatusActive && member.ExpireTime.After(now) {
			if order.Level > member.Level {
				// 升级会员：用户通过补差价对现有会员等级进行升级，不累加有效期，保持原到期时间不变
				expireTime = member.ExpireTime
			} else {
				// 同级续费或降级购买：在原有到期时间基础上累加
				expireTime = member.ExpireTime.Add(time.Duration(order.DurationDays) * 24 * time.Hour)
			}
		} else {
			// 从现在起算
			expireTime = now.Add(time.Duration(order.DurationDays) * 24 * time.Hour)
		}

		// d. 开通或升级会员
		newMember := &model.Member{
			UserID:     order.UserID,
			Level:      order.Level,
			ExpireTime: expireTime,
			Status:     types.MemberStatusActive,
			CreateTime: now,
			UpdateTime: now,
		}
		if err := memberModel.UpsertMember(l.ctx, newMember); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		l.Errorf("[PayCallback] transaction err: %v, orderSn: %s", err, in.OrderSn)
		return nil, err
	}

	// 4. 清除 Redis 缓存
	key := fmt.Sprintf("biz#member#info#%d", order.UserID)
	if _, err := l.svcCtx.BizRedis.DelCtx(l.ctx, key); err != nil {
		l.Errorf("[PayCallback] redis del err: %v, key: %s", err, key)
	}

	return &pb.PayCallbackResponse{}, nil
}
