package myapp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// User stores all field of user
// Used for database
type User struct {
	ID        uint64
	Username  string
	LastLogin time.Time
	Passworld string
}

// UserInfo stores all field For communidate with Odoo
type LoginRequest struct {
	Username string `json:"login"`
	Password string `json:"password"`
}

// LoginResult store info which receive from server odoo
type LoginResult struct {
	Token string `json:"token"`
}

// DoesAnyUserExist returns true if any users exists in the db
// and false otherwise
func (db *DB) DoesAnyUserExist() bool {
	res := false
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		if b != nil {
			stats := b.Stats()
			res = stats.KeyN > 0
		}

		return nil
	})
	if err != nil {
		return false
	}

	return res
}

// GetUser return a user. If no user exists, a gowrapper.ErrNoRows is returned.
func (db *DB) GetUser(username string) (*User, error) {
	var user User
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exits", string(usersBucket))
		}
		userJSON := b.Get([]byte(username))
		if userJSON == nil {
			return ErrNoRows
		}
		return json.Unmarshal(userJSON, &user)
	})
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser creates a user.
// It expects a valid user and password
// It also returns an error if a duplicate user is found.
func (db *DB) CreateUser(username string) (*User, error) {
	// Don't add with config time expire
	//var dateExpire = time.Hour * 24 * time.Duration(viper.GetInt("expireTime"))
	lastLogin := TimeNow()
	user := &User{}

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(usersBucket))
		}
		val := b.Get([]byte(username))
		if val != nil {
			return ErrDuplicateRow
		}

		id, err := b.NextSequence()
		if err != nil {
			return err
		}

		user = &User{ID: id, Username: username, LastLogin: lastLogin}
		userJSON, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("error with marshalling new user object: %s", err)
		}
		return b.Put([]byte(username), userJSON)
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

// UpdateUser return nil. If user is updated return that user
// otherwise return error
func (db *DB) UpdateUser(user *User) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(usersBucket))
		}
		rJSON, err := json.Marshal(user)
		if err != nil {
			return fmt.Errorf("error with marshalling user struct: %s", err)
		}
		return b.Put([]byte(user.Username), rJSON)
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteUser deletes a user. If no user is deleted. nothing happens
// We will return a nil or error
func (db *DB) DeleteUser(username string) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(usersBucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(usersBucket))
		}
		return b.Delete([]byte(username))
	})
	return err
}

// UserDB contain all interface of user
type UserDB interface {
	DoesAnyUserExist() bool
	GetUser(username string) (*User, error)
	CreateUser(username string) (*User, error)
	UpdateUser(user *User) error
	DeleteUser(username string) error
}

// UserSessionDB contain all func of User with Session
type UserSessionDB interface {
	GetUser(username string) (*User, error)
	CreateUser(username string) (*User, error)
	UpdateUser(user *User) error
	CreateSession(token string, uid uint64) (*Session, error)
	UpdateSession(session *Session) error
	GetSessionByUID(uid uint64) (*Session, error)
}
