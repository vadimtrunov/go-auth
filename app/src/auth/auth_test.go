package auth

import (
	"go-auth/store"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite

	user *store.User
}

func (suite *AuthTestSuite) SetupSuite() {
	if err := store.OpenDatabase("../data/teststore.db"); err != nil {
		suite.FailNow("Can't connect to DB", err)
	}
}

func (suite *AuthTestSuite) SetupTest() {
	store.CreateDefaultBacket()

	suite.user = &store.User{
		Email:     "jhondoe@testmail.com",
		Password:  "!strongPwd",
		Nickname:  "JD",
		FirstName: "Jhon",
		LastName:  "Doe",
	}
	suite.user.Create()
}

func (suite *AuthTestSuite) TearDownSuite() {
	store.CloseDatabase()
}

func (suite *AuthTestSuite) TearDownTest() {
	store.DropDatabase()
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

func (suite *AuthTestSuite) TestCreateAuth_WithValidData() {
	creds := Credentials{
		Email:    "jhondoe@testmail.com",
		Password: "!strongPwd",
	}

	ok, _ := creds.Create()
	suite.True(ok)
}

func (suite *AuthTestSuite) TestCreateAuth_WithInvalidEmail() {
	creds := Credentials{
		Email:    "not-jhondoe@testmail.com",
		Password: "!strongPwd",
	}

	ok, errors := creds.Create()
	suite.False(ok)
	suite.Contains(errors, "email")
}

func (suite *AuthTestSuite) TestCreateAuth_WithInvalidCredentials() {
	var creds Credentials

	ok, errors := creds.Create()
	suite.False(ok)
	suite.Contains(errors, "email")
	suite.Contains(errors, "password")
}

func (suite *AuthTestSuite) TestCreateAuth_WithInvalidPassword() {
	creds := Credentials{
		Email:    "jhondoe@testmail.com",
		Password: "!invalidStrongPwd",
	}

	ok, errors := creds.Create()
	suite.False(ok)
	suite.Contains(errors, "password")
}

func (suite *AuthTestSuite) TestAuthorize_WithValidData() {
	creds := Credentials{
		Email:    "jhondoe@testmail.com",
		Password: "!strongPwd",
	}

	creds.Create()

	tokens, _ := creds.Authorize()

	claim := &Claim{}
	tknAuth, _ := jwt.ParseWithClaims(tokens.AuthToken, claim, func(token *jwt.Token) (interface{}, error) {
		return []byte("jwt_secret_key"), nil
	})
	suite.True(tknAuth.Valid)

	suite.Equal(creds.Email, claim.Email)
}

func (suite *AuthTestSuite) TestAuthorize_WithInvalidData() {
	creds := Credentials{
		Email:    "notjhondoe@testmail.com",
		Password: "!strongPwd",
	}

	creds.Create()

	_, err := creds.Authorize()

	suite.NotNil(err)
}

func (suite *AuthTestSuite) TestAuthorize_WithSecondAuthorize() {
	creds := Credentials{
		Email:    "jhondoe@testmail.com",
		Password: "!strongPwd",
	}

	creds.Create()

	oldTokens, err := creds.Authorize()

	suite.Nil(err)

	claim := &Claim{}

	newCreds := Credentials{
		Email:    "jhondoe@testmail.com",
		Password: "!strongPwd",
	}
	newCreds.Create()
	tokens, err := newCreds.Authorize()

	suite.Nil(err)

	tknAuth, _ := jwt.ParseWithClaims(tokens.AuthToken, claim, func(token *jwt.Token) (interface{}, error) {
		return []byte("jwt_secret_key"), nil
	})
	suite.True(tknAuth.Valid)

	suite.Equal(creds.Email, claim.Email)
	suite.Equal(oldTokens, tokens)
}
