package tiktok

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/nandanurseptama/golang-crawler/crawler"
)

type TiktokCrawler interface {
	Search(searchParam string) ([]*VideoCard, error)
}
type Tiktok struct {
	config crawler.Config
}

func NewCrawler(config crawler.Config) TiktokCrawler {
	return &Tiktok{
		config: config,
	}
}

func (crawler *Tiktok) Search(searchParam string) ([]*VideoCard, error) {
	var cards []*VideoCard
	// Chrome options
	opts, err := crawler.config.GetOpts()
	if err != nil {
		return []*VideoCard{}, err
	}

	// Allocator
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Browser context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	uri, err := url.Parse(`https://tiktok.com/search`)
	if err != nil {
		return cards, nil
	}

	query := uri.Query()
	query.Add("q", searchParam)
	query.Add("t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	uri.RawQuery = query.Encode()
	// run task list
	err = chromedp.Run(ctx,
		chromedp.Navigate(uri.String()),
		chromedp.WaitVisible(`[data-e2e="search_top-item-list"]`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var nodes []*cdp.Node
			err := chromedp.Nodes(`[id^="column-item-video-container-"]`, &nodes, chromedp.ByQueryAll).Do(ctx)
			if err != nil {
				return err
			}

			for _, n := range nodes {
				var url, views, caption, user, userLink string

				// Video URL
				_ = chromedp.AttributeValue(fmt.Sprintf("#%s a", n.AttributeValue("id")), "href", &url, nil, chromedp.ByQuery).Do(ctx)

				// Views
				_ = chromedp.Text(fmt.Sprintf("#%s [data-e2e='video-views']", n.AttributeValue("id")), &views, chromedp.ByQuery).Do(ctx)

				// Caption
				_ = chromedp.Text(fmt.Sprintf("#%s [data-e2e='search-card-video-caption'] span", n.AttributeValue("id")), &caption, chromedp.ByQuery).Do(ctx)

				// User name
				_ = chromedp.Text(fmt.Sprintf("#%s [data-e2e='search-card-user-unique-id']", n.AttributeValue("id")), &user, chromedp.ByQuery).Do(ctx)

				// User link
				_ = chromedp.AttributeValue(fmt.Sprintf("#%s [data-e2e='search-card-user-link']", n.AttributeValue("id")), "href", &userLink, nil, chromedp.ByQuery).Do(ctx)

				getTags := func(ctx context.Context) []string {
					tagsSelector := fmt.Sprintf("#%s [data-e2e='search-common-link']", n.AttributeValue("id"))
					var tagNodes []*cdp.Node
					err := chromedp.Nodes(tagsSelector, &tagNodes, chromedp.ByQueryAll).Do(ctx)
					if err != nil {
						return []string{}
					}
					var tags []string
					for _, tagNode := range tagNodes {
						var s string
						_ = chromedp.Text("strong", &s, chromedp.ByQuery, chromedp.FromNode(tagNode)).Do(ctx)

						if s == "" {
							continue
						}

						tags = append(tags, strings.ToLower(strings.ReplaceAll(s, "#", "")))

					}
					return tags
				}

				tagCtx, cancel := context.WithTimeout(ctx, time.Second*2)
				defer cancel()
				tags := getTags(tagCtx)

				cards = append(cards, &VideoCard{
					URL:      url,
					Views:    views,
					Caption:  caption,
					User:     user,
					UserLink: userLink,
					Tags:     tags,
				})
			}
			return nil
		}),
	)

	if err != nil {
		return cards, err
	}
	return cards, err
}
