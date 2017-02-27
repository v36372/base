package myapp

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

// Session contain all field in dabase
type Session struct {
	ID uint64
	//Session     string
	Token       string
	CreatedDate time.Time
	UID         uint64
	Active      bool
}

// CreateSession creates a new session for a user
// One user can have many sessions
func (db *DB) CreateSession(token string, uid uint64) (*Session, error) {
	timeCreated := TimeNow()
	session := &Session{}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(sessionsBucket))
		}
		/*   val := b.Get([]byte(username))*/
		//if val != nil {
		//return ErrDuplicateRow
		/*}*/

		id, err := b.NextSequence()
		if err != nil {
			return err
		}

		session = &Session{ID: id, Token: token,
			CreatedDate: timeCreated, UID: uid, Active: true}
		sessionJSON, err := json.Marshal(session)
		if err != nil {
			return fmt.Errorf("error with marshalling new user object: %s", err)
		}
		byteID := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteID, uint64(id))
		return b.Put(byteID, sessionJSON)
	})
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetSession return a session. If no session exists, a gowrapper.
// ErrNoRows is returned.
func (db *DB) GetSession(id uint64) (*Session, error) {
	var session Session
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exits", string(sessionsBucket))
		}
		byteID := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteID, uint64(id))
		sessionJSON := b.Get([]byte(byteID))
		if sessionJSON == nil {
			return ErrNoRows
		}
		return json.Unmarshal(sessionJSON, &session)
	})
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionByUID return a session. If no session exists, a gowrapper.
// ErrNoRows is returned.
func (db *DB) GetSessionByUID(uid uint64) (*Session, error) {
	var session Session
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exits", string(sessionsBucket))
		}
		byteID := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteID, uint64(uid))
		sessionJSON := b.Get(byteID)
		if sessionJSON == nil {
			return ErrNoRows
		}
		return json.Unmarshal(sessionJSON, &session)
	})
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession return nil. If user is updated return that session
// otherwise return error
func (db *DB) UpdateSession(session *Session) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(sessionsBucket))
		}
		sessionJSON, err := json.Marshal(session)
		if err != nil {
			return fmt.Errorf("error with marshalling session struct: %s", err)
		}
		byteID := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteID, uint64(session.ID))
		return b.Put(byteID, sessionJSON)
	})
	if err != nil {
		return err
	}
	return nil
}

// DeleteSession deletes a session. If no session is deleted. nothing happens
// We will return a nil or error
func (db *DB) DeleteSession(id uint64) error {
	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(sessionsBucket)
		if b == nil {
			return fmt.Errorf("no %s bucket exists", string(sessionsBucket))
		}
		byteID := make([]byte, 8)
		binary.LittleEndian.PutUint64(byteID, uint64(id))
		return b.Delete(byteID)
	})
	if err != nil {
		return err
	}
	return nil
}

// SessionDB contain all inteface of session
type SessionDB interface {
	CreateSession(token string, uid uint64) (*Session, error)
	DeleteSession(id uint64) error
	GetSession(id uint64) (*Session, error)
	GetSessionByUID(uid uint64) (*Session, error)
	UpdateSession(session *Session) error
}
