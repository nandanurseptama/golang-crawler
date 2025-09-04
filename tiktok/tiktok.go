// Copyright The Golang Crawler Author
// SPDX-License-Identifier: Apache-2.0

package tiktok

import (
	"context"

	"github.com/nandanurseptama/golang-crawler/crawler"
)

type TiktokCrawler interface {
	// Search content by search parameter
	Search(ctx context.Context, param SearchParam) ([]ContentItemResp, error)
	// Search user by search parameter
	SearchUser(ctx context.Context, param SearchParam) ([]UserInfoResp, error)
}
type Tiktok struct {
	config crawler.Config
}

func NewCrawler(config crawler.Config) TiktokCrawler {
	return &Tiktok{
		config: config,
	}
}
