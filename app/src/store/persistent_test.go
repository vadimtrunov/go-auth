package store

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type RegistrationTestSuite struct {
	suite.Suite
}

func (suite *RegistrationTestSuite) SetupSuite() {
	if err := OpenDatabase("../data/teststore.db"); err != nil {
		suite.FailNow("Can't connect to DB", err)
	}
}

func (suite *RegistrationTestSuite) SetupTest() {
	CreateDefaultBacket()
}

func (suite *RegistrationTestSuite) TearDownSuite() {
	CloseDatabase()
}

func (suite *RegistrationTestSuite) TearDownTest() {
	DropDatabase()
}

func TestRunSuite(t *testing.T) {
	suite.Run(t, new(RegistrationTestSuite))
}

func (suite *RegistrationTestSuite) TestUserSave_WithValidParams() {
	user := User{
		Email:     "jhondoe@testmail.com",
		Password:  "!strongPwd",
		Nickname:  "JD",
		FirstName: "Jhon",
		LastName:  "Doe",
	}

	ok, _, _ := user.Create()
	suite.True(ok)
}

func (suite *RegistrationTestSuite) TestUserSave_Twice() {
	user := User{
		Email:     "jhondoe@testmail.com",
		Password:  "!strongPwd",
		Nickname:  "JD",
		FirstName: "Jhon",
		LastName:  "Doe",
	}

	ok, _, _ := user.Create()
	suite.True(ok)

	ok, _, _ = user.Create()
	suite.False(ok)
}

func (suite *RegistrationTestSuite) TestUserSave_CheckInDatabase() {
	user := User{
		Email:     "jhondoe@testmail.com",
		Password:  "!strongPwd",
		Nickname:  "JD",
		FirstName: "Jhon",
		LastName:  "Doe",
	}

	ok, _, _ := user.Create()
	suite.True(ok)

	found, savedUser := GetUserByEmail(user.Email)

	suite.True(found)

	suite.Equal(user.Email, savedUser.Email)
	suite.Equal(user.Nickname, savedUser.Nickname)
	suite.Equal(user.FirstName, savedUser.FirstName)
	suite.Equal(user.LastName, savedUser.LastName)
	suite.NotEqual(user.Password, savedUser.HashedPwd)

	suite.Empty(savedUser.Password)
}

func (suite *RegistrationTestSuite) TestUserCreate_WithInvalidParams() {
	user := User{
		Email:     "invalid",
		Password:  "",
		Nickname:  "!",
		FirstName: "!",
		LastName:  "!",
	}
	ok, validationErrors, _ := user.Create()
	suite.False(ok)
	suite.Contains(validationErrors, "email")
	suite.Contains(validationErrors, "password")
	suite.Contains(validationErrors, "nickname")
	suite.Contains(validationErrors, "first_name")
	suite.Contains(validationErrors, "last_name")
}

func (suite *RegistrationTestSuite) TestUserCreate_WithTryToSetHashedPwd() {
	user := User{
		Email:     "jhondoe@testmail.com",
		Password:  "!strongPwd",
		HashedPwd: "cracked!!!",
		Nickname:  "JD",
		FirstName: "Jhon",
		LastName:  "Doe",
	}
	ok, _, _ := user.Create()
	suite.True(ok)
	fmt.Println(user.HashedPwd)
	suite.NotEqual("cracked!!!", user.HashedPwd)
	suite.NotEqual("!strongPwd", user.HashedPwd)
	suite.Empty(user.Password)
}
