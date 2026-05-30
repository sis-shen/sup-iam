package test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/iamapiserver"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/model"
	"github.com/sis-shen/sup-iam/internal/iam-api-server/v1/repository"
	servicemock "github.com/sis-shen/sup-iam/internal/iam-api-server/v1/service/mock"
	"github.com/sis-shen/sup-iam/internal/pkg/log"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserAPI_ApiV1UsersGet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		queryParams    string
		setupMock      func(*servicemock.MockUserCaseInterface)
		expectedStatus int
	}{
		{
			name:        "成功获取用户列表",
			queryParams: "?page=1&page_size=20",
			setupMock: func(m *servicemock.MockUserCaseInterface) {
				m.EXPECT().
					GetUserList(gomock.Any(), gomock.Any()).
					Return(repository.PageResult[*model.User]{
						Items: []*model.User{
							{ID: 1, Username: "alice"},
							{ID: 2, Username: "bob"},
						},
						Total: 2,
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "Service层返回错误",
			queryParams: "?page=1&page_size=20",
			setupMock: func(m *servicemock.MockUserCaseInterface) {
				m.EXPECT().
					GetUserList(gomock.Any(), gomock.Any()).
					Return(repository.PageResult[*model.User]{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建 mock controller
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 创建 mock service
			mockUserCase := servicemock.NewMockUserCaseInterface(ctrl)
			tt.setupMock(mockUserCase)

			// 创建 API 实例
			opts := log.NewOptions()
			opts.Level = "debug"
			if err := opts.Validate(); len(err) != 0 {
				t.Fatal("logger opts.Validate() error")
			}
			logger := log.New(opts)
			api := iamapiserver.NewUserAPI(mockUserCase, logger)

			// 创建请求
			req := httptest.NewRequest("GET", "/api/v1/users"+tt.queryParams, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// 执行
			api.ApiV1UsersGet(c)

			// 断言
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
