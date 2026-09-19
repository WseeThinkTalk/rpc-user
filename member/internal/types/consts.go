package types

const (
	DefaultPageSize = 20
	MaxPageSize     = 50
)

const (
	MemberLevelNormal = 0
	MemberLevelGold   = 1
	MemberLevelDiamond = 2
)

const (
	MemberStatusExpired = 0
	MemberStatusActive  = 1
)

const (
	OrderStatusPending  = 0
	OrderStatusPaid     = 1
	OrderStatusRefunded = 2
)

var MemberLevelNames = map[int32]string{
	MemberLevelNormal:  "普通用户",
	MemberLevelGold:    "黄金会员",
	MemberLevelDiamond: "钻石会员",
}

var MemberRights = map[int32][]string{
	MemberLevelNormal:  {},
	MemberLevelGold:    {"daily_article_10", "daily_qa_5", "remove_ads"},
	MemberLevelDiamond: {"daily_article_unlimited", "daily_qa_unlimited", "remove_ads", "vip_badge", "priority_support"},
}
