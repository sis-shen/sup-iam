package analytics

import "sync"

type AnalyticFilters struct {
	Usernames        []string `json:"usernames" mapstructure:"usernames"`
	SkippedUsernames []string `json:"skippedUsernames" mapstructure:"skippedUsernames"`

	// 内部缓存，不序列化
	once               sync.Once
	usernameSet        map[string]bool
	skippedUsernameSet map[string]bool
}

// ShouldFilter ShouldFilter 适配了黑名单/白名单/ 黑名单> 白名单 的过滤模式
func (a *AnalyticFilters) ShouldFilter(username string) bool {
	a.once.Do(func() {
		a.usernameSet = make(map[string]bool, len(a.Usernames))
		for _, name := range a.Usernames {
			a.usernameSet[name] = true
		}

		a.skippedUsernameSet = make(map[string]bool, len(a.SkippedUsernames))
		for _, name := range a.SkippedUsernames {
			a.skippedUsernameSet[name] = true
		}
	})

	if a.SkippedUsernames != nil && len(a.SkippedUsernames) > 0 && a.skippedUsernameSet[username] {
		return true
	}
	if a.Usernames != nil && len(a.Usernames) > 0 && !a.usernameSet[username] {
		return true
	}

	//没有黑白名单或者不需要过滤，则默认不需要过滤
	return false
}

func (a *AnalyticFilters) HasFilter() bool {
	if a.Usernames != nil && len(a.Usernames) > 0 {
		return true
	}
	if a.SkippedUsernames != nil && len(a.SkippedUsernames) > 0 {
		return true
	}
	return false
}
