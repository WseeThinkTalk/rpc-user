package userlogic

import (
	"context"

	model "rpc-user/internal/model/user"
	"rpc-user/internal/svc"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type FindByMobileLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFindByMobileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FindByMobileLogic {
	return &FindByMobileLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FindByMobileLogic) FindByMobile(in *user.FindByMobileRequest) (*user.FindByMobileResponse, error) {
	u, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)
	if err != nil {
		if err == model.ErrNotFound {
			return &user.FindByMobileResponse{
				Code: 200,
				Msg:  "success",
				Data: nil,
			}, nil
		}
		logx.Errorf("FindByMobile mobile: %s error: %v", in.Mobile, err)
		return nil, err
	}

	return &user.FindByMobileResponse{
		Code: 200,
		Msg:  "success",
		Data: &user.UserAuthData{
			UserId:   int64(u.Id),
			Username: u.Username,
			Avatar:   u.Avatar,
			Password: u.Password,
			Role:     u.Role,
		},
	}, nil
}
