package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func (crawler *Youtube) GetContentComments(ctx context.Context, param SearchContentParam) ([]CommentItem, error) {
	var results []CommentItem
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

	uri, err := url.Parse(`https://youtube.com/watch`)
	if err != nil {
		return results, nil
	}
	query := uri.Query()
	query.Add("v", param.Term)
	uri.RawQuery = query.Encode()

	listenErr := make(chan error, 1)
	resultChannel := make(chan CommentItem, 1)
	shouldScrollCh := make(chan bool, 1)
	var totalLoad int = 0
	var wg sync.WaitGroup
	var commentCount uint64
	var scrollHeight uint64
	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range resultChannel {
			commentCount = commentCount + result.ReplyCount + 1
			fmt.Println("commentCount", commentCount)
			results = append(results, result)
		}
	}()

	// listen target
	chromedp.ListenTarget(
		ctx, func(ev any) {
			switch ev := ev.(type) {
			case *fetch.EventRequestPaused:
				// slog.Info("On EventRequestPaused", slog.Any("url", ev.Request.URL))
				if strings.Contains(ev.Request.URL, "youtubei/v1/next") {
					go crawler.collectFromCommentsAPI(
						ctx,
						ev,
						resultChannel,
						listenErr,
						shouldScrollCh,
						&totalLoad,
						int(param.Scroll),
						param.DelayScrollDuration,
					)
				}
			}
		},
	)

	err = chromedp.Run(ctx,
		network.Enable(),
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{
			{
				URLPattern:   "*youtubei/v1/next*",
				RequestStage: fetch.RequestStageResponse,
			},
		}),
		chromedp.Navigate(uri.String()),

		chromedp.ActionFunc(func(ctx context.Context) error {
			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				//slog.Info("listening should scroll")
				for shouldScroll := range shouldScrollCh {
					//slog.Info("onReceive shouldscroll", slog.Any("value", shouldScroll))
					var nodes []*cdp.Node
					err = chromedp.Nodes(`ytd-item-section-renderer div#contents ytd-comment-thread-renderer`, &nodes, chromedp.ByQueryAll).Do(ctx)
					if err != nil {
						listenErr <- fmt.Errorf("youtube crawler err : %w", err)
						return
					}

					if !shouldScroll {
						return
					}

					var res []byte

					chromedp.Evaluate(`document.querySelector("ytd-item-section-renderer div#contents").scrollHeight;`, &scrollHeight).Do(ctx)

					slog.Info("prevScrollHeight", slog.Any("value", scrollHeight))

					err := chromedp.Evaluate(`window.scrollTo(0,document.querySelector("ytd-item-section-renderer div#contents").scrollHeight);`, &res).Do(ctx)
					//slog.Info("tryScrolling")
					if err != nil {
						listenErr <- fmt.Errorf("youtube crawler err : %w", err)
						return
					}

					//slog.Info("scroll finish")
				}
			}()

			err := chromedp.WaitVisible(`ytd-comments`, chromedp.ByQuery).Do(ctx)

			if err != nil {
				return err
			}

			slog.Info("try scroll to ytd-comments element")

			err = chromedp.Evaluate(`window.scrollTo(0,document.querySelector("ytd-comments").scrollHeight);`, nil).Do(ctx)
			if err != nil {
				return err
			}

			slog.Info("success scroll to ytd-comments element")
			wg.Wait()

			return nil
		}),
	)
	close(listenErr)
	close(resultChannel)
	close(shouldScrollCh)

	wg.Wait()

	if err != nil {
		return results, err
	}

	if err := <-listenErr; err != nil {
		return results, err
	}

	return results, nil
}

func (crawler *Youtube) collectFromCommentsAPI(
	ctx context.Context,
	ev *fetch.EventRequestPaused,
	resultCh chan<- CommentItem,
	errCh chan<- error,
	shouldScrollCh chan<- bool,
	totalLoad *int,
	maxScroll int,
	scrollDelayDuration time.Duration,
) {
	slog.Info("collect from API", slog.Any("totalLoad", *totalLoad))
	scroll := func(canNext <-chan bool) {
		time.Sleep(scrollDelayDuration)

		if *totalLoad >= maxScroll || !<-canNext {
			shouldScrollCh <- false
		} else {
			shouldScrollCh <- true
		}

		// regardless
		*totalLoad = *totalLoad + 1
	}

	c := chromedp.FromContext(ctx)
	e := cdp.WithExecutor(ctx, c.Target)
	bByte, err := fetch.GetResponseBody(ev.RequestID).Do(e)
	slog.Info("get response body succes", slog.Any("contentLength", len(bByte)))
	canNextCh := make(chan bool, 1)
	defer close(canNextCh)
	defer scroll(canNextCh)
	// essential for trigger WaitVisible
	if err != nil {
		errCh <- fmt.Errorf("youtube crawler err : %w", err)
		canNextCh <- false
		return
	}
	slog.Info("try continue response")
	err = fetch.ContinueResponse(ev.RequestID).Do(e)
	slog.Info("continue response finish")
	if err != nil {
		errCh <- fmt.Errorf("youtube crawler err : %w", err)
		canNextCh <- false
		return
	}
	var searchResp GetContentCommentsApiResp
	err = json.Unmarshal(bByte, &searchResp)
	if err != nil {
		errCh <- fmt.Errorf("youtube crawler err : %w", err)
		canNextCh <- false
		return
	}

	results := searchResp.toCommentsItem()
	slog.Info("load commands", slog.Any("commandsLength", len(results)))
	for _, item := range results {
		resultCh <- item
	}

	trigger := searchResp.GetTriggerContinuation()
	canNext := trigger != ""
	canNextCh <- canNext

	slog.Info("finish collect from API", slog.Any("trigger", trigger))
}
