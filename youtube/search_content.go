// Copyright The Golang Crawler Author
// SPDX-License-Identifier: Apache-2.0

package youtube

import (
	"context"
	"net/url"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func (crawler *Youtube) SearchContent(ctx context.Context, param SearchContentParam) ([]VideoItem, error) {
	var results []VideoItem
	// Chrome options
	opts, err := crawler.config.GetOpts()
	if err != nil {
		return results, err
	}

	// Allocator
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Browser context
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	uri, err := url.Parse(`https://youtube.com/results`)
	if err != nil {
		return results, nil
	}
	query := uri.Query()
	query.Add("search_query", param.Term)
	uri.RawQuery = query.Encode()
	err = chromedp.Run(ctx,
		network.Enable(),
		chromedp.Navigate(uri.String()),

		chromedp.ActionFunc(func(ctx context.Context) error {
			err := chromedp.WaitVisible(`ytd-section-list-renderer div#contents ytd-item-section-renderer`, chromedp.ByQuery).Do(ctx)

			if err != nil {
				return err
			}

			var ytInitialData YtInitialDataResp
			err = chromedp.Evaluate(`ytInitialData`, &ytInitialData).Do(ctx)

			if err != nil {
				return err
			}

			results = ytInitialData.GetVideoItems()
			return nil
		}),
	)

	if err != nil {
		return results, err
	}

	return results, nil
}
