// Copyright The Golang Crawler Author
// SPDX-License-Identifier: Apache-2.0

package youtube

import (
	"context"

	"github.com/nandanurseptama/golang-crawler/crawler"
)

type YoutubeCrawler interface {
	SearchContent(ctx context.Context, param SearchContentParam) ([]VideoItem, error)
	SearchChannel(ctx context.Context, param SearchContentParam) ([]ChannelItem, error)
	GetUserContent(ctx context.Context, param SearchContentParam) ([]UserContentItem, error)
}

type Youtube struct {
	config crawler.Config
}

func NewCrawler(config crawler.Config) YoutubeCrawler {
	return &Youtube{
		config: config,
	}
}
