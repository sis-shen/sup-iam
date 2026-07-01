package service

//go:generate mockgen -destination=./mock/auth_case_mock.go -package=mock . AuthCaseInterface
import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casbin/casbin/v2"
	casbinmodel "github.com/casbin/casbin/v2/model"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/analytics"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/load/cache"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/pkg/keys"
	"sync"
	"time"
)

type AuthCaseInterface interface {
	BuildCanonicalString(accessKey string, method string, path string, contentHash, timeStamp string) string
	VerifySecretKey(accessKey string, canonicalString string, signature string) (bool, *model.CachedSecret, error)
	Authorize(secretID string, username string, path string, method string) (bool, []string, error)
}

type AuthCase struct {
	keys          keys.KeysInterface
	enforcerPool  *sync.Pool
	analytics     *analytics.Analytics
	cache         *cache.Cache
	enforcerCache *EnforcerCache
}

func NewAuthCase(ch *cache.Cache, keys keys.KeysInterface, analytics *analytics.Analytics) *AuthCase {
	pool := &sync.Pool{
		New: func() interface{} {
			m, err := casbinmodel.NewModelFromString(CurrenCasbinModelString)
			if err != nil {
				panic(err)
			}
			e, err := casbin.NewEnforcer(m)
			if err != nil {
				//不应该发生
				panic(err)
			}
			return e
		},
	}
	return &AuthCase{
		cache:         ch,
		keys:          keys,
		enforcerPool:  pool,
		analytics:     analytics,
		enforcerCache: NewEnforcerCache(time.Second*5, pool),
	}
}

var _ AuthCaseInterface = (*AuthCase)(nil)

func (ac *AuthCase) BuildCanonicalString(accessKey string, method string, path string, contentHash, timeStamp string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s", accessKey, method, path, contentHash, timeStamp)
}

func (ac *AuthCase) VerifySecretKey(accessKey string, canonicalString string, signature string) (bool, *model.CachedSecret, error) {
	secret, err := ac.cache.GetSecretByAK(accessKey)
	if err != nil {
		return false, nil, err
	}
	nowTIme := time.Now()
	if nowTIme.After(secret.ExpiredAt) {
		return false, nil, errors.New("secret expired")
	}
	ok, err := ac.keys.VerifySecretKey(secret.SecretKey, canonicalString, signature)
	return ok, secret, err
}

func (ac *AuthCase) Authorize(secretID string, username string, path string, method string) (bool, []string, error) {
	startTime := time.Now()
	record := &analytics.AnalyticsRecord{
		UserID:   "",
		Username: username,
		SecretID: secretID,
		Resource: path,
		Action:   method,
	}
	policies, err := ac.cache.GetPolicyListBySecretID(secretID)
	if err != nil {
		return false, nil, err
	}
	matchedPolies := make([]string, 0)

	e, ok := ac.enforcerCache.Get(secretID)
	if !ok {
		e = ac.enforcerPool.Get().(*casbin.Enforcer)
		e.ClearPolicy()
		defer e.ClearPolicy()
		var allDecodes [][]string

		for _, policy := range policies {
			var decodes [][]string

			err := json.Unmarshal([]byte(policy.DSL), &decodes)
			if err != nil {
				return false, nil, err
			}
			allDecodes = append(allDecodes, decodes...)
		}

		ok, err := e.AddPolicies(allDecodes)

		if err != nil {
			return false, nil, err
		}
		if !ok {
			return false, nil, errors.New("failed to add policy")
		}

		ac.enforcerCache.Set(secretID, e)
	}

	ok, err = e.Enforce(username, path, method)

	if err != nil {
		record.Timestamp = time.Now()
		record.Effect = "deny"
		record.Reason = fmt.Sprintf("internal error : %v", err)
		record.LatencyMs = time.Since(startTime).Milliseconds()
		_ = ac.analytics.RecordHit(record)
		return false, nil, err
	}
	if ok {
		record.Timestamp = time.Now()
		record.Effect = "allow"
		record.Reason = fmt.Sprintf("matched policies : %v", matchedPolies)
		record.LatencyMs = time.Since(startTime).Milliseconds()
		_ = ac.analytics.RecordHit(record)
		return true, matchedPolies, nil
	}

	record.Timestamp = time.Now()
	record.Effect = "deny"
	record.Reason = fmt.Sprintf("no policy allowed")
	record.LatencyMs = time.Since(startTime).Milliseconds()
	_ = ac.analytics.RecordHit(record)
	return false, nil, nil
}
