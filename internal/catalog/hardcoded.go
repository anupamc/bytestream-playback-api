package catalog

import "github.com/anupamc/bytestream-playback-api/internal/domain"

type Hardcoded struct {
	baseURL string
	items   map[int64]domain.VideoMeta
}

func NewHardcoded(baseURL string) *Hardcoded {
	return &Hardcoded{
		baseURL: baseURL,
		items: map[int64]domain.VideoMeta{
			46325: {VideoID: 46325, Title: "Example Video 001", StdFilename: "example001", PremiumFilename: "example001-premium", PlaybackExt: ".mp4"},
			46326: {VideoID: 46326, Title: "Example Video 002", StdFilename: "example002", PremiumFilename: "example002-premium", PlaybackExt: ".mp4"},
		},
	}
}

func (c *Hardcoded) Get(id int64) (domain.VideoMeta, bool) { v, ok := c.items[id]; return v, ok }
func (c *Hardcoded) Len() int                              { return len(c.items) }
func (c *Hardcoded) BaseURL() string                       { return c.baseURL }
