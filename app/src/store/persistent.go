package store

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"golang.org/x/crypto/bcrypt"
)

const userBucket = "Users"
const renewTokensBucket = "RenewTokens"
const cryptingCost = 12

var database *bolt.DB

func init() {
	govalidator.TagMap["unique"] = govalidator.Validator(func(email string) bool {
		return !userExists(email)
	})
}

//OpenDatabase opens connection to the persistent DB
func OpenDatabase(store string) error {
	db, err := bolt.Open(store, 0600, nil)
	if err != nil {
		return err
	}
	database = db

	return CreateDefaultBacket()
}

//CreateDefaultBacket create default backet for correct DB work
func CreateDefaultBacket() error {
	return database.Update(func(tx *bolt.Tx) error {

		if _, err := tx.CreateBucketIfNotExists([]byte(userBucket)); err != nil {
			return err
		}

		if _, err := tx.CreateBucketIfNotExists([]byte(renewTokensBucket)); err != nil {
			return err
		}

		return nil
	})
}

//CloseDatabase closes connection and release all resources
func CloseDatabase() {
	database.Close()
}

//DropDatabase cleare all data form database
func DropDatabase() error {
	return database.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(userBucket)); err != nil {
			return err
		}

		if err := tx.DeleteBucket([]byte(renewTokensBucket)); err != nil {
			return err
		}

		return nil
	})
}

//User is datastruct for user with credentials
type User struct {
	Email     string `json:"email" valid:"email,required,unique"`
	Password  string `json:"password" valid:"stringlength(6|64),required"`
	HashedPwd string `json:"hashed_pwd"`
	Nickname  string `json:"nickname" valid:"stringlength(2|100)"`
	FirstName string `json:"first_name" valid:"stringlength(2|100)"`
	LastName  string `json:"last_name" valid:"stringlength(2|100)"`

	validationErrors map[string]string
}

//Create is a method for create user into the store
func (user *User) Create() (valid bool, validationErrors map[string]string, err error) {
	if valid, err = govalidator.ValidateStruct(user); !valid {
		validationErrors = govalidator.ErrorsByField(err)
		return
	}

	cryptedPwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), cryptingCost)
	if err != nil {
		log.Println(err)
		return true, nil, err
	}
	user.HashedPwd = string(cryptedPwd)
	user.Password = ""

	data, err := json.Marshal(user)
	if err != nil {
		return
	}

	err = database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))

		return b.Put([]byte(user.Email), data)
	})
	return
}

//GetUserByEmail get user by email
func GetUserByEmail(email string) (bool, *User) {
	var user User
	var found bool
	err := database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))
		data := b.Get([]byte(email))
		if data == nil {
			return nil
		}
		found = true
		if err := json.Unmarshal(data, &user); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Println("Error while getting user by email: ", err.Error())
	}
	return found, &user
}

//AddRenewToken adds renew token to database
func AddRenewToken(token string, expireAt int64) error {
	err := database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(renewTokensBucket))

		binaries := make([]byte, 8)
		binary.LittleEndian.PutUint64(binaries, uint64(expireAt))

		return b.Put([]byte(token), binaries)
	})
	return err
}

//DeleteRenewToken delete renew token from database
func DeleteRenewToken(token string) error {
	err := database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(renewTokensBucket))

		return b.Delete([]byte(token))
	})
	return err
}

//RenewToken structure with base token data
type RenewToken struct {
	token    string
	expireAt int64
}

//GetAllRenewTokens returns all renew tokens
func GetAllRenewTokens() []RenewToken {
	tokens := make([]RenewToken, 0, 10)
	now := time.Now().Unix()
	database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(renewTokensBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {

			t := RenewToken{
				token:    string(k),
				expireAt: int64(binary.LittleEndian.Uint64(v)),
			}

			if now < t.expireAt {
				tokens = append(tokens, t)
			}
		}
		return nil
	})
	return tokens
}

//ClearRenewTokens deletes all expired tokens from database
func ClearRenewTokens() error {
	now := time.Now().Unix()
	return database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(renewTokensBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)

			if now > int64(binary.LittleEndian.Uint64(v)) {
				if err := b.Delete(k); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func userExists(email string) bool {
	var exist bool
	database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))
		data := b.Get([]byte(email))
		exist = data != nil
		return nil
	})
	return exist
}
