package httpsrv

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthorizationNegative(t *testing.T) {
	router, _, _ := prepareRouter(t)

	h := hData{
		testUsername: "admin",
		testPassword: "invalid",
		router:       router,
	}

	cases := []struct {
		name               string
		method             string
		route              string
		expectedStatusCode int
	}{
		{
			"create_user", http.MethodPost, "/api/v1/manage/user", http.StatusUnauthorized,
		},
		{
			"delete_user", http.MethodDelete, "/api/v1/manage/user/1", http.StatusUnauthorized,
		},
		{
			"create_task", http.MethodPost, "/api/v1/task/", http.StatusUnauthorized,
		},
		{
			"get_task", http.MethodGet, "/api/v1/task/2", http.StatusUnauthorized,
		},
		{
			"update_task", http.MethodPut, "/api/v1/task/3", http.StatusUnauthorized,
		},
		{
			"delete_task", http.MethodDelete, "/api/v1/task/4", http.StatusUnauthorized,
		},
		{
			"get_tasks", http.MethodGet, "/api/v1/tasks/", http.StatusUnauthorized,
		},
		{
			"delete_tasks", http.MethodDelete, "/api/v1/tasks/", http.StatusUnauthorized,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			req, err := http.NewRequest(tt.method, tt.route, nil)
			require.NoError(t, err)

			h.router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatusCode, w.Code)
		})
	}
}
