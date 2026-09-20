package memberlogic

import (
	"context"
	"time"

	"rpc-user/internal/svc"
	types "rpc-user/internal/types/member"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckMemberRightLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckMemberRightLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckMemberRightLogic {
	return &CheckMemberRightLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *CheckMemberRightLogic) CheckMemberRight(in *user.CheckMemberRightRequest) (*user.CheckMemberRightResponse, error) {
	if in.UserId == 0 {
		return nil, code.MemberUserIdEmpty
	}

	m, err := l.svcCtx.MemberModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("[CheckMemberRight] FindByUserId err: %v userId: %d", err, in.UserId)
		return nil, err
	}

	if m == nil || m.Status != types.MemberStatusActive || m.ExpireTime.Before(time.Now()) {
		return &user.CheckMemberRightResponse{
			Code: 200,
			Msg:  "success",
			Data: &user.CheckMemberRightData{HasRight: false, Level: types.MemberLevelNormal},
		}, nil
	}

	if in.RightKey == "" {
		return &user.CheckMemberRightResponse{
			Code: 200,
			Msg:  "success",
			Data: &user.CheckMemberRightData{HasRight: true, Level: m.Level},
		}, nil
	}

	rights := types.MemberRights[m.Level]
	for _, r := range rights {
		if r == in.RightKey {
			return &user.CheckMemberRightResponse{
				Code: 200,
				Msg:  "success",
				Data: &user.CheckMemberRightData{HasRight: true, Level: m.Level},
			}, nil
		}
	}

	return &user.CheckMemberRightResponse{
		Code: 200,
		Msg:  "success",
		Data: &user.CheckMemberRightData{HasRight: false, Level: m.Level},
	}, nil
}
