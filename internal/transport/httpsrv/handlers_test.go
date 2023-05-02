package httpsrv

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"taskmanager/internal/app"
	"taskmanager/internal/config"
	"taskmanager/internal/handler"
	"taskmanager/internal/model"
)

type postgresTest struct {
	userID int
	err    error
}

func (p postgresTest) DeleteUser(ctx context.Context, userID int) error {
	return p.err
}

func (p postgresTest) CreateTask(ctx context.Context, username, password, title string) (int, error) {
	return p.userID, p.err
}

func (p postgresTest) GetTask(ctx context.Context, username, password string, taskID int) (model.Task, error) {
	return model.Task{}, p.err
}

func (p postgresTest) GetTasks(ctx context.Context, username, password string) ([]model.Task, error) {
	return nil, p.err
}

func (p postgresTest) UpdateTask(ctx context.Context, username, password string, taskID int, setValues []string) error {
	return p.err
}

func (p postgresTest) DeleteTask(ctx context.Context, username, password string, taskID int) error {
	return p.err
}

func (p postgresTest) DeleteTasks(ctx context.Context, username, password string) (int64, error) {
	return int64(p.userID), p.err
}

func (p postgresTest) CreateNewUser(ctx context.Context, username string, password string) (int, error) {
	return p.userID, p.err
}

func TestV1CreateUser(t *testing.T) {
	conf, err := config.GetFromFile("../../../configs/conf.toml")
	require.NoError(t, err)

	cases := []struct {
		name              string
		postgres          postgresTest
		username          string
		password          string
		method            string
		route             string
		body              string
		expectedCode      int
		expectedHTTPError handler.HTTPError
	}{
		{
			name: "too short username",
			postgres: postgresTest{
				userID: 0,
			},
			username:     conf.Server.ManageUsername,
			password:     conf.Server.ManagePassword,
			method:       http.MethodPost,
			route:        "/api/v1/manage/user",
			body:         `{"username": "q", "password": "qwerty"}`,
			expectedCode: http.StatusBadRequest,
			expectedHTTPError: handler.HTTPError{
				Type: "USERNAME_REQUIRED",
			},
		},
		{
			name: "too short_password",
			postgres: postgresTest{
				userID: 0,
			},
			username:     conf.Server.ManageUsername,
			password:     conf.Server.ManagePassword,
			method:       http.MethodPost,
			route:        "/api/v1/manage/user",
			body:         `{"username": "qwerty", "password": "qw"}`,
			expectedCode: http.StatusBadRequest,
			expectedHTTPError: handler.HTTPError{
				Type: "PASSWORD_REQUIRED",
			},
		},
		{
			name: "username_missing",
			postgres: postgresTest{
				userID: 0,
			},
			username:     conf.Server.ManageUsername,
			password:     conf.Server.ManagePassword,
			method:       http.MethodPost,
			route:        "/api/v1/manage/user",
			body:         `{"username": "", "password": "qwerty"}`,
			expectedCode: http.StatusBadRequest,
			expectedHTTPError: handler.HTTPError{
				Type: "PARAMETERS_REQUIRED",
			},
		},
		{
			name: "password_missing",
			postgres: postgresTest{
				userID: 0,
			},
			username:     conf.Server.ManageUsername,
			password:     conf.Server.ManagePassword,
			method:       http.MethodPost,
			route:        "/api/v1/manage/user",
			body:         `{"username": "qwerty", "password": ""}`,
			expectedCode: http.StatusBadRequest,
			expectedHTTPError: handler.HTTPError{
				Type: "PARAMETERS_REQUIRED",
			},
		},
		{
			name: "username_exists",
			postgres: postgresTest{
				userID: 0,
				err:    model.ErrUserAlreadyExists,
			},
			username:     conf.Server.ManageUsername,
			password:     conf.Server.ManagePassword,
			method:       http.MethodPost,
			route:        "/api/v1/manage/user",
			body:         `{"username": "qwerty", "password": "qwerty"}`,
			expectedCode: http.StatusBadRequest,
			expectedHTTPError: handler.HTTPError{
				Type: "USERNAME_ALREADY_EXISTS",
			},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			router := testHandlersPrepareRouter(tt.postgres, conf.Server.ManageUsername, conf.Server.ManagePassword)
			w := httptest.NewRecorder()

			req, err := http.NewRequest(tt.method, tt.route, strings.NewReader(tt.body))
			req.SetBasicAuth(tt.username, tt.password)
			require.NoError(t, err)

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code)
			assert.Equal(t, true, strings.Contains(w.Body.String(), tt.expectedHTTPError.Type))
		})
	}
}

func testHandlersPrepareRouter(postgres postgresTest, manageUsername, managePassword string) *gin.Engine {
	serverConf := Conf{
		ManageUsername: manageUsername,
		ManagePassword: managePassword,
	}

	gin.SetMode(gin.TestMode)

	router := gin.New()

	metrics := app.Metrics{
		MetricsRoute: "/metrics",
	}

	serverConf.setRouters(context.Background(), postgres, router, metrics)

	return router
}
