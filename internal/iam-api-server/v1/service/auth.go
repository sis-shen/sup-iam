package service

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/cache"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	"github.com/sis-shen/sup-iam/internal/pkg/jwt"
	"strconv"

	"time"
)

type AuthCaseInterface interface {
	Register(ctx *gin.Context, user *model.User) (*model.User, error)
	Login(ctx *gin.Context, user *model.User) (*model.User, string, string, error) // 先access token后refresh token
	Logout(ctx *gin.Context) error
	RefreshToken(ctx *gin.Context, userID, userName string) (string, string, error) // 先access token后refresh token
}

type AuthCase struct {
	repo      repository.UserRepository
	hasher    PasswordHasherInterface
	jwt       jwt.Manager
	blackList cache.TokenBlackListInterface
}

func NewAuthCase(repo repository.UserRepository, hasher PasswordHasherInterface, jwt jwt.Manager, blackList cache.TokenBlackListInterface) *AuthCase {
	return &AuthCase{repo, hasher, jwt, blackList}
}

func (ac *AuthCase) Register(ctx *gin.Context, user *model.User) (*model.User, error) {
	// 1. 检查用户名是否已存在
	exists, err := ac.repo.ExistsByUsername(ctx, user.Username)
	if err != nil {
		return nil, fmt.Errorf("check username exists failed: %w", err)
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// 2. 检查邮箱是否已存在
	if user.Email != nil {
		exists, err = ac.repo.ExistsByEmail(ctx, *user.Email)
		if err != nil {
			return nil, fmt.Errorf("check email exists failed: %w", err)
		}
		if exists {
			return nil, errors.New("email already registered")
		}
	}

	if user.Phone != nil {
		exists, err = ac.repo.ExistsByPhone(ctx, *user.Phone)
		if err != nil {
			return nil, fmt.Errorf("check phone exists failed: %w", err)
		}
		if exists {
			return nil, errors.New("phone already registered")
		}
	}
	if err := ac.validatePasswordStrength(user.PasswordHash); err != nil {
		return nil, err
	}

	user.PasswordHash, err = ac.hasher.HashPassword(user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("hash user password failed: %w", err)
	}
	user.IsEnable = 1
	user.IsAdmin = 0 //正常注册的用户固定非管理员
	nowTime := time.Now()
	user.CreatedAt = nowTime
	user.UpdatedAt = nowTime

	user.InstanceID = GenerateInstanceID()

	res, err := ac.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("create user failed: %w", err)
	}
	return res, nil
}

// validatePasswordStrength 密码强度校验
func (ac *AuthCase) validatePasswordStrength(password string) error {
	// 长度检查已在 binding 中完成
	// 可添加更多规则：包含数字、大小写、特殊字符等
	hasNumber := false
	hasUpper := false
	hasLower := false

	for _, char := range password {
		switch {
		case char >= '0' && char <= '9':
			hasNumber = true
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		}
	}

	if !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if len(password) < 8 {
		return errors.New("password must contain at least 8 characters")
	}
	return nil
}

func (ac *AuthCase) Login(ctx *gin.Context, user *model.User) (*model.User, string, string, error) {
	// user的PasswordHash字段存储明文密码
	var dbUser *model.User
	var err error
	if user.Username != "" {
		dbUser, err = ac.repo.GetByUsername(ctx, user.Username)
		if err != nil {
			return nil, "", "", fmt.Errorf("get user by username failed: %w", err)
		}
	} else if user.Email != nil {
		dbUser, err = ac.repo.GetByEmail(ctx, *user.Email)
		if err != nil {
			return nil, "", "", fmt.Errorf("get user by email failed: %w", err)
		}
	} else if user.Phone != nil {
		dbUser, err = ac.repo.GetByPhone(ctx, *user.Phone)
		if err != nil {
			return nil, "", "", fmt.Errorf("get user by phone failed: %w", err)
		}
	}

	if dbUser == nil {
		return nil, "", "", errors.New("user not found")
	}

	if ac.hasher.VerifyPassword(user.PasswordHash, dbUser.PasswordHash) != nil {
		return nil, "", "", errors.New("invalid password")
	} else {
		accessToken, refreshToken, err := ac.jwt.GenerateToken(strconv.FormatUint(dbUser.ID, 10), dbUser.Username)
		if err != nil {
			return nil, "", "", fmt.Errorf("generate token failed: %w", err)
		}
		return dbUser, accessToken, refreshToken, nil
	}
}

func (ac *AuthCase) Logout(ctx *gin.Context) error {
	token := ac.jwt.ExtractToken(ctx)
	if token == "" {
		return errors.New("fail to extract token")
	}
	err := ac.blackList.Add(ctx, token)
	if err != nil {
		return err
	}
	return nil
}

func (ac *AuthCase) RefreshToken(ctx *gin.Context, userID string, userName string) (string, string, error) {
	token := ac.jwt.ExtractToken(ctx)
	if token == "" {
		return "", "", errors.New("fail to extract token")
	}
	return ac.jwt.GenerateToken(userID, userName)
}
