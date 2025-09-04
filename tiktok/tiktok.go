package tiktok

import (
	"github.com/nandanurseptama/golang-crawler/crawler"
)

type TiktokCrawler interface {
	// Search content by search parameter
	Search(param SearchParam) ([]ContentItemResp, error)
}
type Tiktok struct {
	config crawler.Config
}

func NewCrawler(config crawler.Config) TiktokCrawler {
	return &Tiktok{
		config: config,
	}
}
