// Copyright The Golang Crawler Author
// SPDX-License-Identifier: Apache-2.0

package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func (crawler *Tiktok) GetUserContent(ctx context.Context, param SearchParam) ([]ContentItemResp, error) {
	var cards []ContentItemResp
	// Chrome options
	opts, err := crawler.config.GetOpts()
	if err != nil {
		return cards, err
	}

	// Allocator
	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	// Browser context
	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	uri, err := url.Parse(fmt.Sprintf(`https://tiktok.com/@%s`, param.Term))
	if err != nil {
		return cards, nil
	}

	listenErr := make(chan error, 1)
	resultChannel := make(chan ContentItemResp, 1)
	shouldScrollCh := make(chan bool, 1)
	var totalLoad int = 0
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range resultChannel {
			cards = append(cards, result)
		}
	}()

	// // run task list
	chromedp.ListenTarget(
		ctx, func(ev any) {
			switch ev := ev.(type) {
			case *fetch.EventRequestPaused:
				go crawler.collectUserContent(ctx, ev, resultChannel, listenErr, shouldScrollCh, &totalLoad, int(param.Scroll))
			}
		},
	)

	err = chromedp.Run(ctx,
		network.Enable(),
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{
			{
				URLPattern:   "*/post/item_list/*",
				RequestStage: fetch.RequestStageResponse,
			},
		}),
		chromedp.Navigate(uri.String()),
		chromedp.WaitVisible(`[data-e2e="user-post-item-list"]`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if param.Scroll < 1 {
				return nil
			}

			for shouldScroll := range shouldScrollCh {
				var nodes []*cdp.Node
				err = chromedp.Nodes(`[id^="column-item-video-container-"]`, &nodes, chromedp.ByQueryAll).Do(ctx)
				if err != nil {
					return fmt.Errorf("tiktok crawler err : %w", err)
				}

				if !shouldScroll {
					return nil
				}
				_, exp, err := runtime.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`).Do(ctx)
				if err != nil {
					return fmt.Errorf("tiktok crawler err : %w", err)
				}
				if exp != nil {
					return fmt.Errorf("tiktok crawler err : %s", exp.Error())
				}
			}

			return nil
		}),
	)

	close(listenErr)
	close(resultChannel)
	close(shouldScrollCh)

	wg.Wait()

	if err != nil {
		return cards, err
	}

	if err := <-listenErr; err != nil {
		return cards, err
	}

	return cards, nil
}

func (crawler *Tiktok) collectUserContent(
	ctx context.Context,
	ev *fetch.EventRequestPaused,
	resultCh chan<- ContentItemResp,
	errCh chan<- error,
	shouldScrollCh chan<- bool,
	totalLoad *int,
	maxScroll int,
) {
	hasMoreCh := make(chan bool, 1)
	scroll := func(canNext <-chan bool) {
		time.Sleep(2 * time.Second)
		// regardless
		*totalLoad = *totalLoad + 1

		if *totalLoad >= maxScroll || !(<-canNext) {
			shouldScrollCh <- false
		} else {
			shouldScrollCh <- true
		}
	}
	defer close(hasMoreCh)

	c := chromedp.FromContext(ctx)
	e := cdp.WithExecutor(ctx, c.Target)
	bByte, err := fetch.GetResponseBody(ev.RequestID).Do(e)
	fetch.ContinueResponse(ev.RequestID).Do(e)
	// essential for trigger WaitVisible
	defer scroll(hasMoreCh)
	if err != nil {
		errCh <- fmt.Errorf("tiktok crawler err : %w", err)
		hasMoreCh <- false
		return
	}
	var searchResp GetUserContentResp
	err = json.Unmarshal(bByte, &searchResp)
	if err != nil {
		errCh <- fmt.Errorf("tiktok crawler err : %w", err)
		hasMoreCh <- false
		return
	}

	for _, v := range searchResp.ItemList {
		resultCh <- v.Item
	}

	hasMoreCh <- (searchResp.HasMore == 1)
}
