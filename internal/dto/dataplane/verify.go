package dataplane

// --------- verify ----------

type VerifyRequest struct {
	AccessKey string `json:"access_key" binding:"required"`
	Method    string `json:"method" binding:"required"`
	Path      string `json:"path" binding:"required"`
	Timestamp int64  `json:"timestamp" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

type VerifyResponse struct {
	Judge      bool   `json:"judge"`
	PolicyInfo string `json:"policy_info"`
}
