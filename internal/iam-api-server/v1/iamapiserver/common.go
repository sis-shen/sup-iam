package iamapiserver

import (
	"errors"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"math"
)

const (
	UserIDKey = "user_id"
	TokenKey  = "token"
)

func parseUserModel(user model.User) User {
	return User{
		Id:           int64(user.ID),
		InstanceId:   user.InstanceID,
		Username:     user.Username,
		Nickname:     user.Nickname,
		IsEnable:     int32(user.IsEnable),
		Phone:        user.Phone,
		Email:        user.Email,
		IsAdmin:      int32(user.IsAdmin),
		ExtendShadow: user.ExtendShadow,
		LoggedAt:     user.LoggedAt,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func getUserModel(user User) model.User {
	return model.User{
		ID:           uint64(user.Id),
		InstanceID:   user.InstanceId,
		Username:     user.Username,
		Nickname:     user.Nickname,
		IsEnable:     uint8(user.IsEnable),
		Phone:        user.Phone,
		Email:        user.Email,
		IsAdmin:      uint8(user.IsAdmin),
		ExtendShadow: user.ExtendShadow,
		LoggedAt:     user.LoggedAt,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

func parseUserModelList(lst []*model.User) ([]User, error) {
	users := make([]User, len(lst))
	for i, user := range lst {
		if user.ID > math.MaxInt64 {
			return nil, errors.New(IntOverFlowError)
		}
		users[i] = parseUserModel(*user)
	}

	return users, nil
}

func ParseSecretModel(secret model.Secret) SecretResponse {
	return SecretResponse{
		Id:          int64(secret.ID),
		InstanceId:  secret.InstanceID,
		UserId:      int64(secret.UserID),
		Username:    secret.Username,
		AccessKey:   secret.AccessKey,
		Description: secret.Description,
		Expires:     &secret.Expires,
		CreatedAt:   secret.CreatedAt,
		UpdatedAt:   secret.UpdatedAt,
	}
}

func ParseSecretModelList(lst []*model.Secret) ([]SecretResponse, error) {
	secrets := make([]SecretResponse, len(lst))
	for i, secret := range lst {
		secrets[i] = ParseSecretModel(*secret)
	}

	return secrets, nil
}

func ParsePolicyModel(policy model.Policy) PolicyResponse {
	return PolicyResponse{
		Id:          int64(policy.ID),
		InstanceId:  policy.InstanceID,
		Name:        policy.Name,
		Username:    policy.Username,
		Description: policy.Description,
		CreatedAt:   policy.CreatedAt,
		UpdatedAt:   policy.UpdatedAt,
		Content:     *policy.PolicyShadow,
	}
}

func ParsePolicyModelList(lst []*model.Policy) ([]PolicyResponse, error) {
	policies := make([]PolicyResponse, len(lst))
	for i, policy := range lst {
		policies[i] = ParsePolicyModel(*policy)
	}
	return policies, nil
}

func ParseBindingModel(binding model.Binding) BindingResponse {
	return BindingResponse{
		BindingId: int64(binding.ID),
		SecretId:  int64(binding.SecretID),
		PolicyId:  int64(binding.PolicyID),
		Username:  binding.Username,
		CreatedAt: binding.CreatedAt,
	}
}

func ParseBindingModelList(lst []*model.Binding) ([]BindingResponse, error) {
	bindings := make([]BindingResponse, len(lst))
	for i, binding := range lst {
		bindings[i] = ParseBindingModel(*binding)
	}
	return bindings, nil
}
