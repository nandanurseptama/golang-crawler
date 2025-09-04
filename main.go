package main

import (
	"fmt"
	"os"

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

	result, err := tiktokCrawler.Search(tiktok.SearchParam{
		Term:   "golang",
		Scroll: 4,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("result len ", len(result))
}
