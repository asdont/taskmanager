package httpsrv

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"taskmanager/internal/app"
	"taskmanager/internal/config"
	"taskmanager/internal/db"
	"taskmanager/internal/model"
	"taskmanager/internal/security"
)

type hData struct {
	testUsername string
	testPassword string
	router       *gin.Engine
}

func TestSimplePositiveScenario(t *testing.T) {
	router, conf, postgres := prepareRouter(t)

	h := hData{
		testUsername: "testuser45983x",
		testPassword: "testpassword45983x",
		router:       router,
	}

	defer clearTestData(t, postgres, h.testUsername, security.SaltPassword(h.testPassword))

	userID := h.createNewUser(t, conf.Server.ManageUsername, conf.Server.ManagePassword)

	taskTitle1 := "45983_1"
	taskID1 := h.createTask(t, taskTitle1)

	h.getTask(t, taskID1, taskTitle1)

	// step 1
	taskNewTitle := "45983_1_test"
	h.updateTaskNewTitle(t, taskID1, taskNewTitle)

	// step 2
	h.checkTaskNewTitle(t, taskID1, taskNewTitle)

	// step 3
	h.updateTaskStatusCompleted(t, taskID1)

	// step 4
	h.checkTaskStatusCompleted(t, taskID1)

	// step 5
	taskTitle2 := "45983_2"
	taskID2 := h.createTask(t, taskTitle2)

	// step 6
	taskTitle3 := "45983_3"
	taskID3 := h.createTask(t, taskTitle3)

	// step 7
	h.deleteTask(t, taskID2)

	// step 8
	h.checkTasks(t, taskID1, taskID3, taskNewTitle, taskTitle3)

	// step 9
	h.deleteTasks(t)

	// step 10
	h.checkNoTasks(t)

	// step 11
	h.deleteUser(t, conf.Server.ManageUsername, conf.Server.ManagePassword, userID)
}

func prepareRouter(t *testing.T) (*gin.Engine, *config.Conf, *sql.DB) {
	var logger *logrus.Logger

	conf, err := config.GetFromFile("../../../configs/conf.toml")
	require.NoError(t, err)

	postgresPool, err := db.Conf{ConnAddress: conf.Postgres.ConnAddress}.CreatePool(logger)
	require.NoError(t, err)

	postgres := model.Postgres{
		Pool:         postgresPool,
		QueryTimeout: conf.Postgres.QueryTimeout,
		Logger:       logger,
	}

	serverConf := Conf{
		ManageUsername: conf.Server.ManageUsername,
		ManagePassword: conf.Server.ManagePassword,
	}

	gin.SetMode(gin.TestMode)

	router := gin.New()

	metrics := app.Metrics{
		MetricsRoute: "/metrics",
	}

	serverConf.setRouters(context.Background(), postgres, router, metrics)

	return router, conf, postgresPool
}

func (h hData) createNewUser(t *testing.T, manageUsername, managePassword string) int {
	reqBody := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: h.testUsername,
		Password: h.testPassword,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPost, "/api/v1/manage/user", bytes.NewReader(body))
	req.SetBasicAuth(manageUsername, managePassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	result := struct {
		UserId int `json:"userId"`
	}{}

	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))

	return result.UserId
}

func (h hData) createTask(t *testing.T, taskName string) int {
	reqBody := struct {
		Title string `json:"title"`
	}{
		Title: taskName,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPost, "/api/v1/task/", bytes.NewReader(body))
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	result := struct {
		TaskID int `json:"taskId"`
	}{}

	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	return result.TaskID
}

func (h hData) getTask(t *testing.T, taskID int, taskName string) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/task/"+strconv.Itoa(taskID), nil)
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result model.Task
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))

	assert.Equal(t, false, result.Status)
	assert.Equal(t, taskName, result.Title)
	assert.Equal(t, result.Created, result.Updated)
	assert.Equal(t, 1, result.Completed.Year())

	return
}

func (h hData) updateTaskNewTitle(t *testing.T, taskID int, taskNewTitle string) {
	reqBody := struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}{
		Title: taskNewTitle,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPut, "/api/v1/task/"+strconv.Itoa(taskID), bytes.NewReader(body))
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	return
}

func (h hData) checkTaskNewTitle(t *testing.T, taskID int, taskNameNew string) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/task/"+strconv.Itoa(taskID), nil)
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result model.Task
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))

	assert.Equal(t, false, result.Status)
	assert.Equal(t, taskNameNew, result.Title)
	assert.Equal(t, true, result.Updated.After(result.Created))
	assert.Equal(t, 1, result.Completed.Year())

	return
}

func (h hData) updateTaskStatusCompleted(t *testing.T, taskID int) {
	reqBody := struct {
		Title     string `json:"title"`
		Completed bool   `json:"completed"`
	}{
		Completed: true,
	}

	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodPut, "/api/v1/task/"+strconv.Itoa(taskID), bytes.NewReader(body))
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	return
}

func (h hData) checkTaskStatusCompleted(t *testing.T, taskID int) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/task/"+strconv.Itoa(taskID), nil)
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result model.Task

	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)

	assert.Equal(t, true, result.Status)
	assert.Equal(t, true, result.Completed.After(result.Updated))

	return
}

func (h hData) deleteTask(t *testing.T, taskID int) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodDelete, "/api/v1/task/"+strconv.Itoa(taskID), nil)
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	return
}

func (h hData) checkTasks(t *testing.T, taskID1, taskID3 int, taskTitle1, taskTitle3 string) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/tasks/", nil)
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []model.Task
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))

	assert.Equal(t, 2, len(result))

	assert.Equal(t, taskTitle1, result[0].Title)
	assert.Equal(t, taskID1, result[0].ID)

	assert.Equal(t, taskTitle3, result[1].Title)
	assert.Equal(t, taskID3, result[1].ID)

	return
}

func (h hData) deleteTasks(t *testing.T) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodDelete, "/api/v1/tasks/", nil)
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	result := struct {
		Quantity int64 `json:"quantity"`
	}{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))

	assert.Equal(t, int64(2), result.Quantity)

	return
}

func (h hData) checkNoTasks(t *testing.T) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/tasks/", nil)
	req.SetBasicAuth(h.testUsername, h.testPassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []model.Task
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &result))

	assert.Equal(t, 0, len(result))

	return
}

func (h hData) deleteUser(t *testing.T, manageUsername, managePassword string, userId int) {
	w := httptest.NewRecorder()

	req, err := http.NewRequest(http.MethodDelete, "/api/v1/manage/user/"+strconv.Itoa(userId), nil)
	req.SetBasicAuth(manageUsername, managePassword)
	require.NoError(t, err)

	h.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	return
}

func clearTestData(t *testing.T, postgres *sql.DB, testUsername, testPassword string) {
	if _, err := postgres.Exec(`
		delete from auth where username = $1 and password = $2
	`,
		testUsername, testPassword,
	); err != nil {
		t.Fatal(err)
	}
}
