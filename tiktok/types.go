package tiktok

type SearchParam struct {
	// search term
	Term string `json:"term"`
	// Total scroll of page
	Scroll uint `json:""`
}

// Content represents extracted info
type Content struct {
	Type int64 `json:"type"`

	URL      string   `json:"url"`
	Views    string   `json:"views"`
	Caption  string   `json:"caption"`
	User     string   `json:"user"`
	UserLink string   `json:"user_link"`
	Tags     []string `json:"tags"`
}

// General Response of api tiktok
type GeneralResp[T any] struct {
	StatusCode int  `json:"status_code"`
	Data       T    `json:"data"`
	HasMore    int `json:"has_more"`
}

// Wrapper of response
type SearchItemResp struct {
	Type int             `json:"type"`
	Item ContentItemResp `json:"item"`
}

type ContentItemResp struct {
	Id           string           `json:"id"`
	Desc         string           `json:"desc"`
	CreateTime   int64            `json:"createTime"`
	Stats        ContentStatsResp `json:"stats"`
	TextExtra    []TextExtraResp  `json:"textExtra"`
	Author       AuthorResp       `json:"author"`
	AuthorStats  AuthorStatsResp  `json:"authorStats"`
	TextLanguage string           `json:"textLanguage"`
}

type ContentStatsResp struct {
	CollectCount uint64 `json:"collectCount"`
	CommentCount uint64 `json:"commentCount"`
	DiggCount    uint64 `json:"diggCount"`
	PlayCount    uint64 `json:"playCount"`
	ShareCount   uint64 `json:"shareCount"`
}

type TextExtraResp struct {
	HastagName string `json:"hashtagName"`
}

type AuthorStatsResp struct {
	FollowerCount  uint64 `json:"followerCount"`
	FollowingCount uint64 `json:"followingCount"`
	DiggCount      uint64 `json:"diggCount"`
	FriendCount    uint64 `json:"friendCount"`
	HeartCount     uint64 `json:"heartCount"`
	VideoCount     uint64 `json:"videoCount"`
}

type AuthorResp struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	// a.k.a username
	UniqueId       string `json:"uniqueId"`
	Signature      string `json:"signature"`
	PrivateAccount bool   `json:"privateAccount"`
	OpenFavorite   bool   `json:"openFavorite"`
}
