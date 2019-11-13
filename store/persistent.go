package store

import (
	"encoding/json"

	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"golang.org/x/crypto/bcrypt"
)

const userBucket = "Users"
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
		_, err := tx.CreateBucketIfNotExists([]byte(userBucket))
		if err != nil {
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
		return tx.DeleteBucket([]byte(userBucket))
	})
}

//User is datastruct for user with credentials
type User struct {
	Email     string `json:"email" valid:"email,required,unique"`
	Password  string `json:"password" valid:"stringlength(6|64),required"`
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
		return valid, nil, err
	}
	user.Password = string(cryptedPwd)

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
