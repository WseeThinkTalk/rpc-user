package code

import "rpc-user/pkg/xcode"

var (
	// User (20000+)
	RegisterNameEmpty = xcode.New(20001, "注册名字不能为空")

	// Member (90000+)
	MemberUserIdEmpty     = xcode.New(90001, "用户ID不能为空")
	LevelInvalid          = xcode.New(90002, "会员等级无效")
	TransactionIdEmpty    = xcode.New(90003, "支付流水号不能为空")
	OrderNotFound         = xcode.New(90004, "订单不存在")
	MemberNotFound        = xcode.New(90005, "会员信息不存在")
	DuplicateTransaction  = xcode.New(90006, "重复的支付流水号")
	OrderSnEmpty          = xcode.New(90007, "订单号不能为空")
	OrderAlreadyProcessed = xcode.New(90008, "订单已被处理，无法重复更新")
	OrderQueryFailed      = xcode.New(90009, "主动查询订单状态失败")

	// Follow (40000+)
	FollowUserIdEmpty   = xcode.New(40001, "关注用户id为空")
	FollowedUserIdEmpty = xcode.New(40002, "被关注用户id为空")
	CannotFollowSelf    = xcode.New(40003, "不能关注自己")
	UserIdEmpty         = xcode.New(40004, "用户id为空")
)
