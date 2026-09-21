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

func (l *CheckMemberRightLogic) CheckMemberRight(in *user.CheckMemberRightRequest) (resp *user.CheckMemberRightResponse, err error) {
	resp = new(user.CheckMemberRightResponse)
	resp.Data = new(user.CheckMemberRightData)

	if in.UserId == 0 {
		resp.Code = int64(code.MemberUserIdEmpty.Code())
		resp.Msg = code.MemberUserIdEmpty.Message()
		return resp, nil
	}

	m, err := l.svcCtx.MemberModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		resp.Code = int64(code.ServerErr.Code())
		resp.Msg = err.Error()
		return resp, nil
	}

	if m == nil || m.Status != types.MemberStatusActive || m.ExpireTime.Before(time.Now()) {
		resp.Data.HasRight = false
		resp.Data.Level = types.MemberLevelNormal
		return resp, nil
	}

	if in.RightKey == "" {
		resp.Data.HasRight = true
		resp.Data.Level = m.Level
		return resp, nil
	}

	rights := types.MemberRights[m.Level]
	// 匹配会员拥有的权限权益
	for _, v := range rights {
		if v == in.RightKey {
			resp.Data.HasRight = true
			resp.Data.Level = m.Level
			return resp, nil
		}
	}

	resp.Data.HasRight = false
	resp.Data.Level = m.Level
	return resp, nil
}
