package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func (crawler *Tiktok) SearchUser(ctx context.Context, param SearchParam) ([]UserInfoResp, error) {
	var cards []UserInfoResp
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

	uri, err := url.Parse(`https://tiktok.com/search/user`)
	if err != nil {
		return cards, nil
	}

	query := uri.Query()
	query.Add("q", param.Term)
	query.Add("t", strconv.FormatInt(time.Now().UnixMilli(), 10))
	uri.RawQuery = query.Encode()
	listenErr := make(chan error, 1)
	resultChannel := make(chan UserInfoResp, 1)
	shouldScrollCh := make(chan bool, 1)
	var totalLoad int = 0
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		defer wg.Done()
		for result := range resultChannel {
			cards = append(cards, result)
		}
	}()

	go func() {
		defer wg.Done()
		for err := range listenErr {
			if err != nil {
				fmt.Println("listen channel err ", err.Error())
			}
		}
	}()

	// // run task list
	chromedp.ListenTarget(
		ctx, func(ev any) {
			switch ev := ev.(type) {
			case *fetch.EventRequestPaused:
				go crawler.collectSearchUserResult(ctx, ev, resultChannel, listenErr, shouldScrollCh, &totalLoad, int(param.Scroll))
			}
		},
	)

	err = chromedp.Run(ctx,
		network.Enable(),
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{
			{
				URLPattern:   "*/search/user/full/*",
				RequestStage: fetch.RequestStageResponse,
			},
		}),
		chromedp.Navigate(uri.String()),
		chromedp.WaitVisible(`[data-e2e="search-user-container"]`, chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if param.Scroll < 1 {
				return nil
			}

			for shouldScroll := range shouldScrollCh {
				var nodes []*cdp.Node
				err = chromedp.Nodes(`[id^="search_user-item-user-link-"]`, &nodes, chromedp.ByQueryAll).Do(ctx)
				if err != nil {
					listenErr <- fmt.Errorf("scrolldown err : %w", err)
					continue
				}

				if !shouldScroll {
					return nil
				}
				_, exp, err := runtime.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`).Do(ctx)
				if err != nil {
					listenErr <- fmt.Errorf("scrolldown err : %w", err)
					continue
				}
				if exp != nil {
					listenErr <- fmt.Errorf("scrolldown err : %w", exp)
					continue
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

	return cards, nil
}

func (t *Tiktok) collectSearchUserResult(
	ctx context.Context,
	ev *fetch.EventRequestPaused,
	resultCh chan<- UserInfoResp,
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
	var searchResp SearchUserResp
	err = json.Unmarshal(bByte, &searchResp)
	if err != nil {
		errCh <- fmt.Errorf("tiktok crawler err : %w", err)
		hasMoreCh <- false
		return
	}

	for _, v := range searchResp.UserList {
		resultCh <- v.UserInfo
	}

	errCh <- nil
	hasMoreCh <- (searchResp.HasMore == 1)

}
