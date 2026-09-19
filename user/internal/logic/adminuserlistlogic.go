package logic

import (
	"context"

	"rpc-user/user/internal/svc"
	"rpc-user/user/service"

	"github.com/zeromicro/go-zero/core/logx"
)

type AdminUserListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAdminUserListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminUserListLogic {
	return &AdminUserListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *AdminUserListLogic) AdminUserList(in *service.AdminUserListRequest) (*service.AdminUserListResponse, error) {
	users, err := l.svcCtx.UserModel.FindAll(l.ctx, in.Keyword, in.Cursor, in.PageSize)
	if err != nil {
		return nil, err
	}

	var isEnd bool
	if len(users) < int(in.PageSize) {
		isEnd = true
	}

	var items []*service.AdminUserItem
	var nextCursor int64 = in.Cursor
	for _, u := range users {
		items = append(items, &service.AdminUserItem{
			UserId:       int64(u.Id),
			Username:     u.Username,
			Mobile:       u.Mobile,
			Avatar:       u.Avatar,
			Role:         u.Role,
			DisplayId:    u.DisplayId,
			Bio:          u.Bio,
			Gender:       int32(u.Gender),
			ProfileCover: u.ProfileCover,
			CreateTime:   u.CreateTime.Unix(),
		})
		nextCursor = int64(u.Id)
	}

	return &service.AdminUserListResponse{
		Items:  items,
		Cursor: nextCursor,
		IsEnd:  isEnd,
	}, nil
}
