package tiktok

// VideoCard represents extracted info
type VideoCard struct {
	URL      string   `json:"url"`
	Views    string   `json:"views"`
	Caption  string   `json:"caption"`
	User     string   `json:"user"`
	UserLink string   `json:"user_link"`
	Tags     []string `json:"tags"`
}
