package service

//go:generate mockgen -destination=./mock/auth_case_mock.go -package=mock . AuthCaseInterface
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/casbin/casbin/v2"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-auth-server/v1/rpc"
	"github.com/sis-shen/sup-iam/internal/pkg/keys"
	"strconv"
	"time"
)

type AuthCaseInterface interface {
	BuildCanonicalString(accessKey string, method string, path string, contentHash, timeStamp string) string
	VerifySecretKey(ctx context.Context, accessKey string, canonicalString string, signature string) (bool, *model.Secret, error)
	Authorize(ctx context.Context, secretID string, accessKey string, path string, method string) (bool, []string, error)
}

type AuthCase struct {
	cli      rpc.RpcClientInterface
	keys     keys.Keys
	enforcer casbin.Enforcer
}

func NewAuthCase(cli rpc.RpcClientInterface) *AuthCase {
	return &AuthCase{cli: cli}
}

var _ AuthCaseInterface = (*AuthCase)(nil)

func (ac *AuthCase) BuildCanonicalString(accessKey string, method string, path string, contentHash, timeStamp string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s", accessKey, method, path, contentHash, timeStamp)
}

func (ac *AuthCase) VerifySecretKey(ctx context.Context, accessKey string, canonicalString string, signature string) (bool, *model.Secret, error) {
	secret, err := ac.cli.GetSecretByAK(ctx, accessKey)
	if err != nil {
		return false, nil, err
	}
	nowTIme := time.Now()
	if nowTIme.After(time.Unix(secret.Expires, 0)) {
		return false, nil, errors.New("secret expired")
	}
	ok, err := ac.keys.VerifySecretKey(secret.SecretKey, canonicalString, signature)
	return ok, secret, err
}

func (ac *AuthCase) Authorize(ctx context.Context, secretID string, accessKey string, path string, method string) (bool, []string, error) {
	policies, err := ac.cli.GetPolicyListBySecretID(ctx, secretID)
	if err != nil {
		return false, nil, err
	}
	matchedPolies := make([]string, 0)
	for _, policy := range policies {
		ac.enforcer.ClearPolicy()
		var policies [][]string
		err := json.Unmarshal([]byte(*policy.PolicyShadow), &policies)
		if err != nil {
			return false, nil, err
		}
		ok, err := ac.enforcer.AddPolicy(policies)
		if err != nil {
			return false, nil, err
		}

		if !ok {
			return false, nil, errors.New("failed to add policy")
		}

		ok, err = ac.enforcer.Enforce(accessKey, path, method)
		if err != nil {
			return false, nil, err
		}
		if ok {
			matchedPolies = append(matchedPolies, strconv.FormatUint(policy.ID, 10))
			return true, matchedPolies, nil
		}
	}
	return false, nil, nil
}
