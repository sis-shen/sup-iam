package pumps

import (
	"github.com/sis-shen/sup-iam/internal/iam-pump/analytics"
	"time"
)

type CommonPump struct {
	Filters               analytics.AnalyticFilters
	Timeout               time.Duration
	OmitDetailedRecording bool
}
