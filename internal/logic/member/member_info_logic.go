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

func (l *MemberInfoLogic) MemberInfo(in *user.MemberInfoRequest) (*user.MemberInfoResponse, error) {
	if in.UserId == 0 {
		return nil, code.MemberUserIdEmpty
	}

	m, err := l.svcCtx.MemberModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("[MemberInfo] FindByUserId err: %v userId: %d", err, in.UserId)
		return nil, err
	}
	if m == nil {
		return &user.MemberInfoResponse{
			Code: 200,
			Msg:  "success",
			Data: &user.MemberInfoData{
				UserId:    in.UserId,
				Level:     types.MemberLevelNormal,
				LevelName: types.MemberLevelNames[types.MemberLevelNormal],
				Status:    types.MemberStatusActive,
			},
		}, nil
	}

	if m.Status == types.MemberStatusExpired {
		return &user.MemberInfoResponse{
			Code: 200,
			Msg:  "success",
			Data: &user.MemberInfoData{
				UserId:    in.UserId,
				Level:     types.MemberLevelNormal,
				LevelName: types.MemberLevelNames[types.MemberLevelNormal],
				Status:    types.MemberStatusExpired,
			},
		}, nil
	}

	return &user.MemberInfoResponse{
		Code: 200,
		Msg:  "success",
		Data: &user.MemberInfoData{
			UserId:     m.UserID,
			Level:      m.Level,
			LevelName:  types.MemberLevelNames[m.Level],
			ExpireTime: m.ExpireTime.Unix(),
			Status:     m.Status,
		},
	}, nil
}
