package crawler

import (
	"encoding/json"
	"fmt"

	"github.com/chromedp/chromedp"
)

type Config struct {
	// Chrome path at your local
	ChromePath string `json:"-"`

	Headless           bool `json:"headless"`
	DisableGPU         bool `json:"disable-gpu"`
	NoSandbox          bool `json:"no-sandbox"`
	DisableDevSHMUsage bool `json:"disable-dev-shm-usage"`
}

func (param *Config) GetFlags() (map[string]any, error) {

	cBytes, err := json.Marshal(param)

	if err != nil {
		return map[string]any{}, fmt.Errorf("config param err : %w", err)
	}

	flags := map[string]any{}
	json.Unmarshal(cBytes, &flags)

	return flags, nil
}
func (param *Config) GetOpts() ([]chromedp.ExecAllocatorOption, error) {
	defaultOpts := []chromedp.ExecAllocatorOption{chromedp.ExecPath(param.ChromePath)}

	flags, err := param.GetFlags()
	if err != nil {
		return []chromedp.ExecAllocatorOption{}, fmt.Errorf("failed to get chromedp opts : %w", err)
	}
	for k, v := range flags {
		defaultOpts = append(defaultOpts, chromedp.Flag(k, v))
	}

	return defaultOpts, nil
}
