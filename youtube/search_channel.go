package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func (crawler *Youtube) SearchChannel(ctx context.Context, param SearchContentParam) ([]ChannelItem, error) {
	var results []ChannelItem
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
	// flag filter by channel when search at youtube
	query.Add("sp", "EgIQAg%3D%3D")
	uri.RawQuery = query.Encode()

	listenErr := make(chan error, 1)
	resultChannel := make(chan ChannelItem, 1)
	shouldScrollCh := make(chan bool, 1)
	var totalLoad int = 0
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for result := range resultChannel {
			results = append(results, result)
		}
	}()

	// listen target
	chromedp.ListenTarget(
		ctx, func(ev any) {
			switch ev := ev.(type) {
			case *fetch.EventRequestPaused:
				// slog.Info("On EventRequestPaused", slog.Any("url", ev.Request.URL))
				if strings.Contains(ev.Request.URL, "youtubei/v1/search") {
					go crawler.collectFromSearchChannelAPI(
						ctx, ev, resultChannel,
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
				URLPattern:   "*youtubei/v1/search*",
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
					err = chromedp.Nodes(`ytd-section-list-renderer div#contents ytd-item-section-renderer`, &nodes, chromedp.ByQueryAll).Do(ctx)
					if err != nil {
						listenErr <- fmt.Errorf("youtube crawler err : %w", err)
						return
					}

					if !shouldScroll {
						return
					}

					var res []byte

					err := chromedp.Evaluate(`window.scrollTo(0,document.querySelector("ytd-section-list-renderer div#contents").scrollHeight);`, &res).Do(ctx)
					//slog.Info("tryScrolling")
					if err != nil {
						listenErr <- fmt.Errorf("youtube crawler err : %w", err)
						return
					}

					//slog.Info("scroll finish")
				}
			}()

			err := chromedp.WaitVisible(`ytd-section-list-renderer div#contents ytd-item-section-renderer`, chromedp.ByQuery).Do(ctx)

			if err != nil {
				return err
			}

			results, err = crawler.collectChannelFromSearchInitialData(ctx)
			//slog.Info("collect initial data finish")

			if err != nil {
				return err
			}
			//slog.Info("get initial data success", slog.Any("resultLength", len(results)))

			time.Sleep(param.DelayScrollDuration)
			shouldScrollCh <- true
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

// collect initial data from youtube page
func (crawler *Youtube) collectChannelFromSearchInitialData(ctx context.Context) ([]ChannelItem, error) {
	var ytInitialData YtInitialDataResp
	err := chromedp.Evaluate(`ytInitialData`, &ytInitialData).Do(ctx)

	if err != nil {
		return []ChannelItem{}, err
	}

	return ytInitialData.GetChannelItems(), nil
}

func (crawler *Youtube) collectFromSearchChannelAPI(
	ctx context.Context,
	ev *fetch.EventRequestPaused,
	resultCh chan<- ChannelItem,
	errCh chan<- error,
	shouldScrollCh chan<- bool,
	totalLoad *int,
	maxScroll int,
	scrollDelayDuration time.Duration,
) {
	//slog.Info("collect from API", slog.Any("totalLoad", *totalLoad))
	scroll := func() {
		time.Sleep(scrollDelayDuration)

		if *totalLoad >= maxScroll {
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
	defer scroll()
	// essential for trigger WaitVisible
	if err != nil {
		errCh <- fmt.Errorf("youtube crawler err : %w", err)
		return
	}
	err = fetch.ContinueResponse(ev.RequestID).Do(e)
	if err != nil {
		errCh <- fmt.Errorf("youtube crawler err : %w", err)
		return
	}
	var searchResp SearchContentResp
	err = json.Unmarshal(bByte, &searchResp)
	if err != nil {
		errCh <- fmt.Errorf("youtube crawler err : %w", err)
		return
	}

	// slog.Info("get response body succes", slog.Any("contentLength", len(bByte)))

	// slog.Info("load commands", slog.Any("commandsLength", len(searchResp.OnResponseReceiveCommands)))
	for _, command := range searchResp.OnResponseReceiveCommands {
		//slog.Info("load continueItem", slog.Any("continueItemLength", len(command.AppendContinuationItemsAction.ContinuationItems)))
		for _, continueItem := range command.AppendContinuationItemsAction.ContinuationItems {
			//slog.Info("load videoItem", slog.Any("videoItemLength", len(continueItem.ItemSectionRenderer.GetVideoItems())))
			for _, videoItem := range continueItem.ItemSectionRenderer.GetChannelItems() {
				resultCh <- videoItem
			}

		}
	}

	// slog.Info("finish collect from API")
}
