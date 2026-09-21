package followlogic

import (
	"context"
	"strconv"
	"time"

	model "rpc-user/internal/model/follow"
	"rpc-user/internal/svc"
	types "rpc-user/internal/types/follow"
	"rpc-user/pkg/code"
	"rpc-user/user"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/threading"
)

const userFansExpireTime = 3600 * 24 * 2

type FansListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFansListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FansListLogic {
	return &FansListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FansListLogic) FansList(in *user.FansListRequest) (resp *user.FansListResponse, err error) {
	resp = new(user.FansListResponse)
	resp.Data = new(user.FansListData)
	resp.Data.Items = make([]*user.FansItem, 0)

	if in.UserId == 0 {
		resp.Code = int64(code.UserIdEmpty.Code())
		resp.Msg = code.UserIdEmpty.Message()
		return resp, nil
	}
	if in.PageSize == 0 {
		in.PageSize = types.DefaultPageSize
	}
	if in.Cursor == 0 {
		in.Cursor = time.Now().Unix()
	}
	var (
		isCache, isEnd bool
		lastId, cursor int64
		fansUserIds    []int64
		fansModel      []*model.Follow
		curPage        []*user.FansItem
	)
	fansUIds, createTime, _ := l.cacheFansUserIds(l.ctx, in.UserId, in.Cursor, in.PageSize)
	if len(fansUIds) > 0 {
		isCache = true
		if fansUIds[len(fansUIds)-1] == -1 {
			fansUIds = fansUIds[:len(fansUIds)-1]
			isEnd = true
		}
		if len(fansUIds) == 0 {
			return resp, nil
		}
		fansModel, err = l.svcCtx.FollowModel.FindByUserIds(l.ctx, in.UserId, fansUIds)
		if err != nil {
			resp.Code = int64(code.ServerErr.Code())
			resp.Msg = err.Error()
			return resp, nil
		}
		// 构建粉丝ID到关注关系模型的映射
		fansMap := make(map[int64]*model.Follow)
		for _, v := range fansModel {
			fansMap[v.UserID] = v
		}
		// 组装缓存中的粉丝列表
		for _, v := range fansUIds {
			fansItem := &user.FansItem{
				UserId: v,
			}
			if f, ok := fansMap[v]; ok {
				fansItem.Id = f.ID
				fansItem.CreateTime = f.CreateTime.Unix()
			}
			curPage = append(curPage, fansItem)
			fansUserIds = append(fansUserIds, v)
		}
	} else {
		fansModel, err = l.svcCtx.FollowModel.FindByFollowedUserId(l.ctx, in.UserId, types.CacheMaxFansCount)
		if err != nil {
			resp.Code = int64(code.ServerErr.Code())
			resp.Msg = err.Error()
			return resp, nil
		}
		if len(fansModel) == 0 {
			return resp, nil
		}
		var firstPageFans []*model.Follow
		if len(fansModel) > int(in.PageSize) {
			firstPageFans = fansModel[:in.PageSize]
		} else {
			firstPageFans = fansModel
			isEnd = true
		}
		// 组装第一页粉丝数据
		for _, v := range firstPageFans {
			curPage = append(curPage, &user.FansItem{
				Id:         v.ID,
				UserId:     v.UserID,
				CreateTime: v.CreateTime.Unix(),
			})
			fansUserIds = append(fansUserIds, v.UserID)
		}
	}
	if len(curPage) > 0 {
		pageLast := curPage[len(curPage)-1]
		lastId = pageLast.Id
		if isCache {
			cursor = createTime
		} else {
			cursor = pageLast.CreateTime
		}
		if cursor < 0 {
			cursor = 0
		}
		// 根据游标匹配当前分页起始位置
		for k, v := range curPage {
			if v.CreateTime == in.Cursor && v.Id == in.Id {
				curPage = curPage[k:]
				break
			}
		}
	}
	fc, err := l.svcCtx.FollowCountModel.FindByUserIds(l.ctx, fansUserIds)
	_ = err
	uidFansCount := make(map[int64]int)
	uidFollowCount := make(map[int64]int)
	// 汇总各用户的粉丝数与关注数
	for _, v := range fc {
		uidFansCount[v.UserID] = v.FansCount
		uidFollowCount[v.UserID] = v.FollowCount
	}
	// 回填各粉丝项的统计数据
	for _, v := range curPage {
		v.FansCount = int64(uidFansCount[v.UserId])
		v.FollowCount = int64(uidFollowCount[v.UserId])
	}
	resp.Data.IsEnd = isEnd
	resp.Data.Cursor = cursor
	resp.Data.Id = lastId
	if curPage != nil {
		resp.Data.Items = curPage
	}

	if !isCache {
		threading.GoSafe(func() {
			if len(fansModel) < types.CacheMaxFansCount && len(fansModel) > 0 {
				fansModel = append(fansModel, &model.Follow{UserID: -1})
			}
			_ = l.addCacheFans(context.Background(), in.UserId, fansModel)
		})
	}

	return resp, nil
}

func (l *FansListLogic) cacheFansUserIds(ctx context.Context, userId, cursor, pageSize int64) ([]int64, int64, error) {
	key := userFansKey(userId)
	b, err := l.svcCtx.BizRedis.ExistsCtx(ctx, key)
	if err == nil && b {
		_ = l.svcCtx.BizRedis.ExpireCtx(ctx, key, userFansExpireTime)
	}
	pairs, err := l.svcCtx.BizRedis.ZrevrangebyscoreWithScoresAndLimitCtx(ctx, key, 0, cursor, 0, int(pageSize))
	if err != nil {
		return nil, 0, err
	}
	var uids []int64
	var score int64
	// 解析 Redis ZSet 元素中的粉丝用户 ID 与时间戳
	for _, v := range pairs {
		uid, err := strconv.ParseInt(v.Key, 10, 64)
		if err != nil {
			continue
		}
		score = v.Score
		uids = append(uids, uid)
	}

	return uids, score, nil
}

func (l *FansListLogic) addCacheFans(ctx context.Context, userId int64, follows []*model.Follow) error {
	if len(follows) == 0 {
		return nil
	}
	key := userFansKey(userId)
	// 批量写入粉丝缓存到 Redis ZSet
	for _, v := range follows {
		var score int64
		if v.UserID == -1 {
			score = 0
		} else {
			score = v.CreateTime.Unix()
		}
		_, err := l.svcCtx.BizRedis.ZaddCtx(ctx, key, score, strconv.FormatInt(v.UserID, 10))
		if err != nil {
			return err
		}
	}

	return l.svcCtx.BizRedis.ExpireCtx(ctx, key, userFansExpireTime)
}
