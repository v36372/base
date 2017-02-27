package main

import (
	"encoding/json"
	"myapp"
	"net/http"
)

// createNewUser will create new user and new session for user
func createUser(loginRequest myapp.LoginRequest,
	usr *myapp.User, db myapp.UserSessionDB) (string, error) {
	// Create new user if on server we don't have this user
	token, err := CreateToken(loginRequest.Username)
	if err != nil {
		return "", err
	}
	usr, err = db.CreateUser(loginRequest.Username)
	if err != nil {
		return "", err
	}

	// Create new token in session
	session, err := db.CreateSession(token, usr.ID)
	if err != nil {
		return "", err
	}
	return session.Token, nil
}

// updateUser will update user and update session belong user.
func updateUser(usr *myapp.User, db myapp.UserSessionDB) (string, error) {
	// Update user
	token, err := CreateToken(usr.Username)
	if err != nil {
		return "", err
	}
	session, err := db.GetSessionByUID(usr.ID)
	session.Token = token
	err = db.UpdateUser(usr)
	if err != nil {
		return "", err
	}

	err = db.UpdateSession(session)
	if err != nil {
		return "", err
	}
	return session.Token, nil
}

// LoginHandler process login of user
func (a *App) LoginHandler(db myapp.UserSessionDB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		// Get info of user from Request
		var loginRequest myapp.LoginRequest
		var resultToken string
		err := json.NewDecoder(req.Body).Decode(&loginRequest)
		if err != nil {
			a.logr.Log("error decode: %s", err)
			return newAPIError(400, "error decode: %s", err)
		}

		// Process begin
		usr, err := db.GetUser(loginRequest.Username)
		if err == nil && usr.Passworld == loginRequest.Password {
			sessionToken, err := updateUser(usr, db)
			if err != nil {
				a.logr.Log("update create user error: %s", err)
				return newAPIError(400, "update user fail: %s", err)
			}
			resultToken = sessionToken
		} else {
			return newAPIError(404, "login fail.", err)
		}

		loginResult := myapp.LoginResult{
			Token: resultToken,
		}
		err = json.NewEncoder(w).Encode(loginResult)
		if err != nil {
			return newAPIError(404, "error when return json ", err)
		}
		return nil
	}
}
