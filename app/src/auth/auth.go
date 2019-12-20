package auth

import (
	"errors"
	"go-auth/session"
	"go-auth/store"
	"log"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("jwt_secret_key")

const authTokenLiveMinutes = 5
const renewTokenLiveMinutes = 60 * 24

//Credentials struct for credentials
type Credentials struct {
	Email     string `json:"email" valid:"required"`
	Password  string `json:"password" valid:"required"`
	isCreated bool
	claim     Claim
}

//Claim stuct contains auth user data
type Claim struct {
	Email      string
	Nickname   string
	FirstName  string
	LastName   string
	AuthToken  string
	RenewToken string

	jwt.StandardClaims
}

//Authorize authorize credentials and returns authorized user jwt token
func (creds *Credentials) Authorize() (*Claim, error) {
	if !creds.isCreated {
		return nil, errors.New("You need create credentilas first using method 'Create'")
	}

	stringifyToken := func(minutes int) (string, error) {
		expiresAt := time.Now().Add(time.Duration(minutes) * time.Minute)
		creds.claim.ExpiresAt = expiresAt.Unix()
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, creds.claim)
		return token.SignedString(jwtKey)
	}

	var err error
	creds.claim.AuthToken, err = stringifyToken(authTokenLiveMinutes)
	if err != nil {
		return nil, err
	}

	creds.claim.RenewToken, err = stringifyToken(renewTokenLiveMinutes)
	if err != nil {
		return nil, err
	}

	s := session.Create(creds.Email, creds.claim.ExpiresAt)
	s.Add(creds.claim.RenewToken)

	return &creds.claim, nil
}

func (creds *Credentials) verifyPassword(hashedPwd string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPwd), []byte(creds.Password))
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

//Create create auth data according provided credentials
func (creds *Credentials) Create() (bool, map[string]string) {
	if valid, err := govalidator.ValidateStruct(creds); !valid {
		return false, govalidator.ErrorsByField(err)
	}

	found, user := store.GetUserByEmail(creds.Email)
	if !found {
		return false, map[string]string{
			"email": "Can't found user with such email",
		}
	}

	if ok := creds.verifyPassword(user.HashedPwd); !ok {
		return false, map[string]string{
			"password": "Invalid password",
		}
	}

	creds.claim.Email = user.Email
	creds.claim.Nickname = user.Nickname
	creds.claim.FirstName = user.FirstName
	creds.claim.LastName = user.LastName

	creds.isCreated = true
	return true, nil
}
