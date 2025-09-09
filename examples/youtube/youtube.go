package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/nandanurseptama/golang-crawler/crawler"
	"github.com/nandanurseptama/golang-crawler/youtube"
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

	youtubeCrawler := youtube.NewCrawler(config)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()

	results, err := youtubeCrawler.GetUserContent(ctx, youtube.SearchContentParam{
		Term:                "mitocw",
		Scroll:              3,
		DelayScrollDuration: time.Second * 3,
	})

	if err != nil {
		fmt.Println("failed get search content", err.Error())
		return
	}

	slog.Info("results", slog.Any("length", len(results)))

	for i, v := range results {
		slog.Info("item at", slog.Any("index", i), slog.Any("value", v))
	}

}
