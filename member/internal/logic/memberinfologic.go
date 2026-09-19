package logic

import (
	"context"

	"rpc-user/member/code"
	"rpc-user/member/internal/svc"
	"rpc-user/member/internal/types"
	"rpc-user/member/pb"

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

func (l *MemberInfoLogic) MemberInfo(in *pb.MemberInfoRequest) (*pb.MemberInfoResponse, error) {
	if in.UserId == 0 {
		return nil, code.UserIdEmpty
	}

	member, err := l.svcCtx.MemberModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("[MemberInfo] FindByUserId err: %v userId: %d", err, in.UserId)
		return nil, err
	}
	if member == nil {
		return &pb.MemberInfoResponse{
			UserId:    in.UserId,
			Level:     types.MemberLevelNormal,
			LevelName: types.MemberLevelNames[types.MemberLevelNormal],
			Status:    types.MemberStatusActive,
		}, nil
	}

	if member.Status == types.MemberStatusExpired {
		return &pb.MemberInfoResponse{
			UserId:    in.UserId,
			Level:     types.MemberLevelNormal,
			LevelName: types.MemberLevelNames[types.MemberLevelNormal],
			Status:    types.MemberStatusExpired,
		}, nil
	}

	return &pb.MemberInfoResponse{
		UserId:     member.UserID,
		Level:      member.Level,
		LevelName:  types.MemberLevelNames[member.Level],
		ExpireTime: member.ExpireTime.Unix(),
		Status:     member.Status,
	}, nil
}
