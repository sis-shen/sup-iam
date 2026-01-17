package dataplane

// --------- verify ----------

type VerifyRequest struct {
	AccessKey   string `json:"access_key" binding:"required"`
	Method      string `json:"method" binding:"required"`
	Path        string `json:"path" binding:"required"`
	ContentHash string `json:"content_hash" binding:"required"`
	Timestamp   int64  `json:"timestamp" binding:"required"`
	Signature   string `json:"signature" binding:"required"`
}

type VerifyResponse struct {
	Allowed         bool     `json:"allowed"`
	Reason          string   `json:"reason,omitempty"`
	HTTPCode        int      `json:"http_code,omitempty"`
	MatchedPolicies []string `json:"matched_policies,omitempty"`
}
