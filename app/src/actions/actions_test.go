package actions

import (
	"bytes"
	"encoding/json"
	"errors"
	"go-auth/auth"
	"go-auth/store"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type mockResponse struct {
	Status string `json:"status"`
}

type HTTPHandler func(http.ResponseWriter, *http.Request)

func mockAction(r *http.Request) (int, interface{}) {
	response := mockResponse{"OK"}
	return http.StatusOK, response
}

func mockInvalidAction(r *http.Request) (int, interface{}) {
	return http.StatusOK, func() {}
}

func mockInternalErrorAction(r *http.Request) (int, interface{}) {
	return http.StatusInternalServerError, errors.New("Some fatal error")
}

func TestRun_WithValidData(t *testing.T) {
	rr := proccedRequest(http.MethodGet, Run(mockAction, http.MethodGet), t)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, rr.Header().Get("Content-Type"), "application/json")

	assert.JSONEq(t, `{"status":"OK"}`, rr.Body.String())
}

func TestRun_WithInvalidHttpMethod(t *testing.T) {
	rr := proccedRequest(http.MethodPost, Run(mockAction, http.MethodPost), t)

	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestRun_WithUnmarshalizedData(t *testing.T) {
	rr := proccedRequest(http.MethodGet, Run(mockInvalidAction, http.MethodGet), t)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	jsonErr := make(map[string]string)
	json.NewDecoder(rr.Body).Decode(&jsonErr)
	assert.Contains(t, jsonErr, "error")

	assert.Equal(t, rr.Header().Get("Content-Type"), "application/json")
}

func TestRun_WithInternalServerError(t *testing.T) {
	rr := proccedRequest(http.MethodGet, Run(mockInternalErrorAction, http.MethodGet), t)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func proccedRequest(method string, httpHandler HTTPHandler, t *testing.T) *httptest.ResponseRecorder {
	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(httpHandler)
	handler.ServeHTTP(rr, req)

	return rr
}

func TestHealthcheck(t *testing.T) {
	assert := assert.New(t)
	status, response := Healthcheck(nil)

	assert.Equal(status, http.StatusOK, "Invalid response status code")

	data := response.(healthCheckResponse)

	assert.Equal(data.Status, "Ok", "Invalid response body. Invalid field 'status'")
	assert.Equal(data.Code, http.StatusOK, "Invalid response body. Invalid field 'code'")
}

type DefaultTestSuit struct {
	suite.Suite
}

type RegistrationTestSuite struct {
	DefaultTestSuit
}

func (suite *DefaultTestSuit) SetupSuite() {
	if err := store.OpenDatabase("../data/teststore.db"); err != nil {
		suite.FailNow("Can't connect to DB", err)
	}
}

func (suite *DefaultTestSuit) SetupTest() {
	store.CreateDefaultBacket()
}

func (suite *DefaultTestSuit) TearDownSuite() {
	store.CloseDatabase()
}

func (suite *DefaultTestSuit) TearDownTest() {
	store.DropDatabase()
}

func TestRunRegistrationSuite(t *testing.T) {
	suite.Run(t, new(RegistrationTestSuite))
}

func (suite *RegistrationTestSuite) TestRegistration_WithValidParams() {

	user := store.User{
		Email:     "jhondoe@testmail.com",
		Password:  "!strongPwd",
		Nickname:  "JD",
		FirstName: "Jhon",
		LastName:  "Doe",
	}

	data, _ := json.Marshal(user)
	request, _ := http.NewRequest(http.MethodPost, "/registration", bytes.NewReader(data))
	status, result := Registration(request)

	suite.Equal(http.StatusCreated, status)
	suite.Equal(result, nil)
}

func (suite *RegistrationTestSuite) TestRegistration_WithInvalidDate() {

	user := store.User{
		Email:     "invalid",
		Password:  "",
		Nickname:  "!",
		FirstName: "!",
		LastName:  "!",
	}

	data, _ := json.Marshal(user)
	request, _ := http.NewRequest(http.MethodPost, "/registration", bytes.NewReader(data))
	status, result := Registration(request)

	suite.Equal(http.StatusUnprocessableEntity, status)
	suite.Contains(result, "email")
	suite.Contains(result, "password")
	suite.Contains(result, "nickname")
	suite.Contains(result, "first_name")
	suite.Contains(result, "last_name")
}

func (suite *RegistrationTestSuite) TestRegistration_WithEmptyParams() {
	request, _ := http.NewRequest(http.MethodPost, "/registration", bytes.NewReader(nil))
	status, _ := Registration(request)

	suite.Equal(http.StatusBadRequest, status)
}

type LoginTestSuite struct {
	DefaultTestSuit

	user *store.User
}

func (suite *LoginTestSuite) SetupTest() {
	suite.DefaultTestSuit.SetupTest()
	suite.user = &store.User{
		Email:     "jhondoe@testmail.com",
		Password:  "!strongPwd",
		Nickname:  "JD",
		FirstName: "Jhon",
		LastName:  "Doe",
	}
	suite.user.Create()
}

func TestRunLoginSuite(t *testing.T) {
	suite.Run(t, new(LoginTestSuite))
}

func (suite *LoginTestSuite) TestLogin_WithValidData() {
	creds := auth.Credentials{
		Email:    "jhondoe@testmail.com",
		Password: "!strongPwd",
	}

	data, _ := json.Marshal(creds)
	request, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(data))
	status, tokens := Login(request)

	tkns := tokens.(*auth.Claim)

	suite.Equal(http.StatusOK, status)
	suite.NotEmpty(tkns.AuthToken)
	suite.NotEmpty(tkns.RenewToken)

	authUser := &auth.Claim{}

	tknAuth, _ := jwt.ParseWithClaims(tkns.AuthToken, authUser, func(token *jwt.Token) (interface{}, error) {
		return []byte("jwt_secret_key"), nil
	})
	suite.True(tknAuth.Valid)
	suite.Equal(suite.user.Email, authUser.Email)
	suite.Equal(suite.user.Nickname, authUser.Nickname)
	suite.Equal(suite.user.FirstName, authUser.FirstName)
	suite.Equal(suite.user.LastName, authUser.LastName)

	tknRenew, _ := jwt.ParseWithClaims(tkns.RenewToken, authUser, func(token *jwt.Token) (interface{}, error) {
		return []byte("jwt_secret_key"), nil
	})

	suite.True(tknRenew.Valid)
}

func (suite *LoginTestSuite) TestLogin_WithInvalidEmail() {
	creds := auth.Credentials{
		Email:    "not_jhondoe@testmail.com",
		Password: "!strongPwd",
	}

	data, _ := json.Marshal(creds)
	request, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(data))
	status, errors := Login(request)

	errMap := errors.(map[string]string)

	suite.Equal(http.StatusUnprocessableEntity, status)
	suite.Contains(errMap, "email")
}

func (suite *LoginTestSuite) TestLogin_WithInvalidPassword() {
	creds := auth.Credentials{
		Email:    "jhondoe@testmail.com",
		Password: "!strongWrongPwd",
	}

	data, _ := json.Marshal(creds)
	request, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(data))
	status, errors := Login(request)

	errMap := errors.(map[string]string)

	suite.Equal(http.StatusUnprocessableEntity, status)
	suite.Contains(errMap, "password")
}

func (suite *LoginTestSuite) TestLogin_WithEmptyData() {
	request, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewReader(nil))
	status, _ := Login(request)

	suite.Equal(http.StatusBadRequest, status)
}