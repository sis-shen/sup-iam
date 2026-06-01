package pumps

import (
	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"time"
)

type CommonPump struct {
	Filters               *analytics.AnalyticFilters
	Timeout               time.Duration
	OmitDetailedRecording bool
}

func (c *CommonPump) GetFilters() *analytics.AnalyticFilters {
	return c.Filters
}
func (c *CommonPump) GetTimeout() time.Duration {
	return c.Timeout
}

func (c *CommonPump) GetOmitDetailEnable() bool {
	return c.OmitDetailedRecording
}

func (c *CommonPump) SetOmitDetailEnable(v bool) {
	c.OmitDetailedRecording = v
}

func (c *CommonPump) SetFilters(filters *analytics.AnalyticFilters) {
	c.Filters = filters
}
func (c *CommonPump) SetTimeout(timeout time.Duration) {
	c.Timeout = timeout
}
