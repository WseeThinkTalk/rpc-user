package logic

import (
	"rpc-user/user/internal/model"
	"context"

	"rpc-user/user/internal/svc"
	"rpc-user/user/service"

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

func (l *FindByMobileLogic) FindByMobile(in *service.FindByMobileRequest) (*service.FindByMobileResponse, error) {
	// 查询用户
	user, err := l.svcCtx.UserModel.FindOneByMobile(l.ctx, in.Mobile)

	if err != nil {
		if err == model.ErrNotFound {
			// 用户不存在，返回空响应（不是错误）
			return &service.FindByMobileResponse{}, nil
		}
		// 其他错误（数据库连接失败等）
		logx.Errorf("FindByMobile mobile: %s error: %v", in.Mobile, err)
		return nil, err
	}

	// 用户存在，返回用户信息
	return &service.FindByMobileResponse{
		UserId:   int64(user.Id),
		Username: user.Username,
		Avatar:   user.Avatar,
		Password: user.Password,
		Role:     user.Role,
	}, nil
}
