package logic

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"rpc-user/member/code"
	"rpc-user/member/internal/model"
	"rpc-user/member/internal/svc"
	"rpc-user/member/internal/types"
	"rpc-user/member/pb"

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

// planPricing defines the pricing table. Frontend only sends plan_id.
var planPricing = map[int32]struct {
	Amount       int64
	DurationDays int32
	Level        int32
}{
	1: {Amount: 5990, DurationDays: 30, Level: types.MemberLevelGold},      // 黄金月卡 59.90
	5: {Amount: 15900, DurationDays: 90, Level: types.MemberLevelGold},     // 黄金季卡 159.00
	2: {Amount: 59900, DurationDays: 365, Level: types.MemberLevelGold},    // 黄金年卡 599.00
	3: {Amount: 9990, DurationDays: 30, Level: types.MemberLevelDiamond},   // 钻石月卡
	6: {Amount: 25900, DurationDays: 90, Level: types.MemberLevelDiamond},  // 钻石季卡
	4: {Amount: 99900, DurationDays: 365, Level: types.MemberLevelDiamond}, // 钻石年卡
}

func generateOrderSn() string {
	n, err := rand.Int(rand.Reader, big.NewInt(99999999))
	if err != nil {
		n = big.NewInt(0)
	}
	return fmt.Sprintf("MEMBER_%s%08d", time.Now().Format("20060102150405"), n.Int64())
}

func (l *CreateOrderLogic) CreateOrder(in *pb.CreateOrderRequest) (*pb.CreateOrderResponse, error) {
	if in.UserId <= 0 {
		return nil, code.UserIdEmpty
	}

	var planId int32 = 0
	if in.Level == 1 { // 黄金
		if in.DurationDays == 30 || in.DurationDays == 1 {
			planId = 1
		} else if in.DurationDays == 90 || in.DurationDays == 3 {
			planId = 5
		} else if in.DurationDays == 365 || in.DurationDays == 12 {
			planId = 2
		}
	} else if in.Level == 2 { // 钻石
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
		return nil, code.LevelInvalid
	}

	// 1. 生成订单号
	orderSn := generateOrderSn()

	// 2. 优惠金额计算：如果开通的是钻石会员且用户当前是有效的黄金会员，则扣减对应时长黄金会员的金额（补差价）
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

	// 3. 插入数据库（金额以最终计算的补差价金额为准）
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

	err := l.svcCtx.MemberOrderModel.Insert(l.ctx, order)
	if err != nil {
		l.Errorf("[CreateOrder] Insert member_order error: %v, order_sn: %s", err, orderSn)
		return nil, err
	}

	return &pb.CreateOrderResponse{
		OrderSn: orderSn,
		Amount:  finalAmount,
	}, nil
}
