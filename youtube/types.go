// Copyright The Golang Crawler Author
// SPDX-License-Identifier: Apache-2.0

package youtube

type SearchContentParam struct {
	Term string
}

// type GeneralResponse struct {
// 	OnResponseReceiveCommands []OnResponseReceiveCommandsResp `json:"onResponseReceivedCommands"`
// }

// type OnResponseReceiveCommandsResp struct {
// 	AppendContinuationItemsAction AppendContinuationItemsActionResp `json:"appendContinuationItemsAction"`
// }

// type AppendContinuationItemsActionResp struct {
// 	ContinuationItems []ContinuationItemResp `json:"continuationItems"`
// }

type VideoItem struct {
	ID         string      `json:"ID"`
	Channel    Channel     `json:"channel"`
	Thumbnails []Thumbnail `json:"thumbnail"`
	// video duration in seconds
	Duration      uint64 `json:"duration"`
	DurationText  string `json:"durationText"`
	ViewCount     uint64 `json:"viewCount"`
	ViewCountText string `json:"viewCountText"`
	Title         string `json:"title"`
	// Video description
	Desc string `json:"desc"`
	// video published time
	PublishedTime string `json:"publishedTime"`
}

type Channel struct {
	Name     string `json:"name"`
	ID       string `json:"ID"`
	Endpoint string `json:"endpoint"`
}

type Thumbnail struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
