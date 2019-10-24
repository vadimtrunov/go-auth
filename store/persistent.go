package store

import (
	"encoding/json"
	"log"

	"github.com/asaskevich/govalidator"
	"github.com/boltdb/bolt"
	"golang.org/x/crypto/bcrypt"
)

const userBucket = "Users"
const cryptingCost = 12

var database *bolt.DB

func init() {
	govalidator.TagMap["unique"] = govalidator.Validator(func(email string) bool {
		var user User
		if found, err := Get(&user, email); found {
			return false
		} else if err != nil {
			log.Println(err)
			return false
		}
		return true
	})
}

//OpenDatabase opens connection to the persistent DB
func OpenDatabase() error {
	db, err := bolt.Open("data/store.db", 0600, nil)
	if err != nil {
		return err
	}
	database = db

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

//User is datastruct for user with credentials
type User struct {
	Email     string `json:"email" valid:"email,required,unique"`
	Password  string `json:"password" valid:"stringlength(6|64),required"`
	Nickname  string `json:"nickname" valid:"stringlength(2|100)"`
	FirstName string `json:"first_name" valid:"stringlength(2|100)"`
	LastName  string `json:"last_name" valid:"stringlength(2|100)"`
}

//Save is a method for create or update user into the store
//You can't change email and password in this method. Use change ChangeCredentials for it
func (user *User) Save() error {
	return database.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))

		var u User
		exists, err := getUserByEmail(&u, user.Email, b)
		if err != nil {
			return err
		}

		if exists {
			user.Password = u.Password
			user.Email = u.Email
		} else {
			cryptedPwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), cryptingCost)
			if err != nil {
				return err
			}
			user.Password = string(cryptedPwd)
		}

		data, err := json.Marshal(user)
		if err != nil {
			return err
		}

		return b.Put([]byte(user.Email), data)
	})
}

//ChangeCredentials changing password and email of existing user
func (user *User) ChangeCredentials() error {
	//TODO implement method
	return nil
}

//Get found user by email
func Get(user *User, email string) (bool, error) {
	var found bool

	err := database.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(userBucket))

		exists, err := getUserByEmail(user, email, b)
		if err != nil {
			return err
		}
		found = exists
		return nil
	})

	return found, err
}

func getUserByEmail(user *User, email string, b *bolt.Bucket) (bool, error) {
	data := b.Get([]byte(email))
	if data == nil {
		return false, nil
	}
	if err := json.Unmarshal(data, user); err != nil {
		return true, err
	}

	return true, nil
}
