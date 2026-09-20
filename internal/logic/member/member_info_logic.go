package memberlogic

import (
	"context"

	"rpc-user/internal/svc"
	types "rpc-user/internal/types/member"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type MemberInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMemberInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MemberInfoLogic {
	return &MemberInfoLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *MemberInfoLogic) MemberInfo(in *user.MemberInfoRequest) (resp *user.MemberInfoResponse, err error) {
	resp = new(user.MemberInfoResponse)
	resp.Code = 200
	resp.Msg = "success"
	resp.Data = new(user.MemberInfoData)

	if in.UserId == 0 {
		return nil, code.MemberUserIdEmpty
	}

	m, err := l.svcCtx.MemberModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("[MemberInfo] FindByUserId err: %v userId: %d", err, in.UserId)
		return nil, err
	}
	if m == nil {
		resp.Data.UserId = in.UserId
		resp.Data.Level = types.MemberLevelNormal
		resp.Data.LevelName = types.MemberLevelNames[types.MemberLevelNormal]
		resp.Data.Status = types.MemberStatusActive
		return resp, nil
	}

	if m.Status == types.MemberStatusExpired {
		resp.Data.UserId = in.UserId
		resp.Data.Level = types.MemberLevelNormal
		resp.Data.LevelName = types.MemberLevelNames[types.MemberLevelNormal]
		resp.Data.Status = types.MemberStatusExpired
		return resp, nil
	}

	resp.Data.UserId = m.UserID
	resp.Data.Level = m.Level
	resp.Data.LevelName = types.MemberLevelNames[m.Level]
	resp.Data.ExpireTime = m.ExpireTime.Unix()
	resp.Data.Status = m.Status

	return resp, nil
}
