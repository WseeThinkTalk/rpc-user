package userlogic

import (
	"context"

	"rpc-user/internal/svc"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpgradePasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpgradePasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpgradePasswordLogic {
	return &UpgradePasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpgradePasswordLogic) UpgradePassword(in *user.UpgradePasswordRequest) (resp *user.UpgradePasswordResponse, err error) {
	resp = new(user.UpgradePasswordResponse)

	if err := l.svcCtx.UserModel.UpdatePassword(l.ctx, uint64(in.UserId), in.PasswordHash); err != nil {
		l.Errorf("[UpgradePassword] err: %v userId: %d", err, in.UserId)
		return nil, err
	}
	return resp, nil
}
