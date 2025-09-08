// Copyright The Golang Crawler Author
// SPDX-License-Identifier: Apache-2.0

package youtube

import "time"

type SearchContentParam struct {
	Term string

	// Total scroll of page
	// it will scroll as much of this
	Scroll uint

	// Delay duration between scroll
	DelayScrollDuration time.Duration
}

type VideoItem struct {
	ID         string      `json:"ID"`
	Channel    Channel     `json:"channel"`
	Thumbnails []Thumbnail `json:"thumbnails"`
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

type ChannelItem struct {
	Channel
	Description         string      `json:"description"`
	SubscriberCountText string      `json:"subsciberCountText"`
	Thumbnails          []Thumbnail `json:"thumbnails"`
}
