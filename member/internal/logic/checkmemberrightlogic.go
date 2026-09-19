package logic

import (
	"context"
	"time"

	"rpc-user/member/code"
	"rpc-user/member/internal/svc"
	"rpc-user/member/internal/types"
	"rpc-user/member/pb"

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

func (l *CheckMemberRightLogic) CheckMemberRight(in *pb.CheckMemberRightRequest) (*pb.CheckMemberRightResponse, error) {
	if in.UserId == 0 {
		return nil, code.UserIdEmpty
	}

	member, err := l.svcCtx.MemberModel.FindByUserId(l.ctx, in.UserId)
	if err != nil {
		l.Errorf("[CheckMemberRight] FindByUserId err: %v userId: %d", err, in.UserId)
		return nil, err
	}

	if member == nil || member.Status != types.MemberStatusActive || member.ExpireTime.Before(time.Now()) {
		return &pb.CheckMemberRightResponse{HasRight: false, Level: types.MemberLevelNormal}, nil
	}

	if in.RightKey == "" {
		return &pb.CheckMemberRightResponse{HasRight: true, Level: member.Level}, nil
	}

	rights := types.MemberRights[member.Level]
	for _, r := range rights {
		if r == in.RightKey {
			return &pb.CheckMemberRightResponse{HasRight: true, Level: member.Level}, nil
		}
	}

	return &pb.CheckMemberRightResponse{HasRight: false, Level: member.Level}, nil
}
