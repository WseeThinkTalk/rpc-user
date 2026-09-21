package memberlogic

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	model "rpc-user/internal/model/member"
	"rpc-user/internal/svc"
	types "rpc-user/internal/types/member"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateOrderLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateOrderLogic {
	return &CreateOrderLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

var planPricing = map[int32]struct {
	Amount       int64
	DurationDays int32
	Level        int32
}{
	1: {Amount: 5990, DurationDays: 30, Level: types.MemberLevelGold},
	5: {Amount: 15900, DurationDays: 90, Level: types.MemberLevelGold},
	2: {Amount: 59900, DurationDays: 365, Level: types.MemberLevelGold},
	3: {Amount: 9990, DurationDays: 30, Level: types.MemberLevelDiamond},
	6: {Amount: 25900, DurationDays: 90, Level: types.MemberLevelDiamond},
	4: {Amount: 99900, DurationDays: 365, Level: types.MemberLevelDiamond},
}

func generateOrderSn() string {
	n, err := rand.Int(rand.Reader, big.NewInt(99999999))
	if err != nil {
		n = big.NewInt(0)
	}
	return fmt.Sprintf("MEMBER_%s%08d", time.Now().Format("20060102150405"), n.Int64())
}

func (l *CreateOrderLogic) CreateOrder(in *user.CreateOrderRequest) (resp *user.CreateOrderResponse, err error) {
	resp = new(user.CreateOrderResponse)
	resp.Data = new(user.CreateOrderData)

	if in.UserId <= 0 {
		resp.Code = int64(code.MemberUserIdEmpty.Code())
		resp.Msg = code.MemberUserIdEmpty.Message()
		resp.Data = nil
		return resp, nil
	}

	var planId int32 = 0
	if in.Level == 1 {
		if in.DurationDays == 30 || in.DurationDays == 1 {
			planId = 1
		} else if in.DurationDays == 90 || in.DurationDays == 3 {
			planId = 5
		} else if in.DurationDays == 365 || in.DurationDays == 12 {
			planId = 2
		}
	} else if in.Level == 2 {
		if in.DurationDays == 30 || in.DurationDays == 1 {
			planId = 3
		} else if in.DurationDays == 90 || in.DurationDays == 3 {
			planId = 6
		} else if in.DurationDays == 365 || in.DurationDays == 12 {
			planId = 4
		}
	}

	plan, ok := planPricing[planId]
	if !ok {
		resp.Code = int64(code.LevelInvalid.Code())
		resp.Msg = code.LevelInvalid.Message()
		resp.Data = nil
		return resp, nil
	}

	orderSn := generateOrderSn()

	finalAmount := plan.Amount
	if plan.Level == types.MemberLevelDiamond {
		currMember, err := l.svcCtx.MemberModel.FindByUserId(l.ctx, in.UserId)
		if err == nil && currMember != nil && currMember.Level == types.MemberLevelGold && currMember.ExpireTime.After(time.Now()) && currMember.Status == types.MemberStatusActive {
			var goldPlanId int32 = 0
			if plan.DurationDays == 30 {
				goldPlanId = 1
			} else if plan.DurationDays == 90 {
				goldPlanId = 5
			} else if plan.DurationDays == 365 {
				goldPlanId = 2
			}
			if goldPlan, ok := planPricing[goldPlanId]; ok {
				if finalAmount > goldPlan.Amount {
					finalAmount -= goldPlan.Amount
				}
			}
		}
	}

	now := time.Now()
	order := &model.MemberOrder{
		UserID:        in.UserId,
		OrderSN:       orderSn,
		Level:         plan.Level,
		DurationDays:  plan.DurationDays,
		Amount:        finalAmount,
		PayChannel:    in.PayChannel,
		TransactionID: "",
		Status:        types.OrderStatusPending,
		CreateTime:    now,
		UpdateTime:    now,
	}

	err = l.svcCtx.MemberOrderModel.Insert(l.ctx, order)
	if err != nil {
		resp.Code = 500
		resp.Msg = err.Error()
		resp.Data = nil
		return resp, nil
	}

	resp.Data.OrderSn = orderSn
	resp.Data.Amount = finalAmount

	return resp, nil
}
