// Copyright The Golang Crawler
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"github.com/nandanurseptama/golang-crawler/crawler"
	"github.com/nandanurseptama/golang-crawler/tiktok"
)

func main() {
	godotenv.Load()
	path := os.Getenv("CHROME_PATH")
	config := crawler.Config{
		ChromePath:         path,
		Headless:           false,
		DisableGPU:         true,
		NoSandbox:          true,
		DisableDevSHMUsage: true,
	}
	flags, _ := config.GetFlags()
	for i, v := range flags {
		fmt.Println(i, v)
	}

	tiktokCrawler := tiktok.NewCrawler(config)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		result, err := tiktokCrawler.SearchUser(ctx, tiktok.SearchParam{
			Term:   "ahmad",
			Scroll: 4,
		})
		if err != nil {
			fmt.Println("failed search user", err.Error())
			return
		}

		fmt.Println("search user result len ", result)
	}()

	go func() {
		defer wg.Done()
		result, err := tiktokCrawler.Search(ctx, tiktok.SearchParam{
			Term:   "golang",
			Scroll: 4,
		})
		if err != nil {
			fmt.Println("failed search content", err.Error())
			return
		}

		fmt.Println("search content result len ", result)
	}()

	wg.Wait()

}
