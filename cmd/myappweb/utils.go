package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"myapp"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/spf13/viper"
)

// Header is part of a token
type Header struct {
	Type string `json:"type"`
	ALG  string `json:"alg"`
}

// Payload is part of a token
type Payload struct {
	Iss  string `json:"iss"`
	EXP  string `json:"exp"`
	Name string `json:"name"`
}

// getUser returns the user from the context object in the request.
func getUser(req *http.Request) *myapp.User {
	if rv := context.Get(req, UserKeyName); rv != nil {
		res := rv.(*myapp.User)
		return res
	}
	return nil
}

// GetParamsObj returns a httprouter params object given the request.
func GetParamsObj(req *http.Request) httprouter.Params {
	ps := context.Get(req, Params).(httprouter.Params)
	return ps
}

// CreateToken creates token, it include: header, payload, signature
func CreateToken(name string) (string, error) {
	secretKey := viper.GetString("secretKey")
	headr := &Header{
		Type: "JWT",
		ALG:  "HS256",
	}
	payld := &Payload{
		Iss:  "epos",
		EXP:  "HS256",
		Name: name,
	}

	headerJSON, _ := json.Marshal(headr)
	payloadJSON, _ := json.Marshal(payld)
	headerString := base64.URLEncoding.EncodeToString(headerJSON)
	payloadString := base64.URLEncoding.EncodeToString(payloadJSON)

	encodeString := headerString + "." + payloadString

	key := []byte(secretKey)
	h := hmac.New(sha256.New, key)
	_, err := h.Write([]byte(encodeString))
	if err != nil {
		return "", err
	}

	signatureString := base64.StdEncoding.EncodeToString(h.Sum(nil))

	return encodeString + "." + signatureString, nil
}

// DecodeHeader use to decode token and get header
func DecodeHeader(token string) (Header, error) {
	tokenArray := strings.Split(token, ".")
	headr := Header{}

	if len(tokenArray) != 3 {
		return headr, fmt.Errorf("wrong token")
	}
	headerByte, err := base64.StdEncoding.DecodeString(tokenArray[0])
	if err != nil {
		return headr, fmt.Errorf("wrong token")
	}
	if err = json.Unmarshal(headerByte, &headr); err != nil {
		return headr, fmt.Errorf("error when parse json DecodeHeader")
	}

	return headr, nil
}

// DecodePayload use to decode token and get payload
func DecodePayload(token string) (Payload, error) {
	tokenArray := strings.Split(token, ".")
	payld := Payload{}

	if len(tokenArray) != 3 {
		return payld, fmt.Errorf("wrong token")
	}
	payloadByte, err := base64.StdEncoding.DecodeString(tokenArray[1])
	if err != nil {
		return payld, err
	}
	if err = json.Unmarshal(payloadByte, &payld); err != nil {
		return payld, fmt.Errorf("error when parse json DecodePayload")
	}

	return payld, nil
}

// SplitDomainString splits the string into 3 parts with separator is math operator in the middle.
func SplitDomainString(strInput string) ([]interface{}, error) {

	// Caution when inputing new operators, since it do a loop contain check from left to right
	operators := []string{"!=", "<=", ">=", "=?", "<", ">", " not ilike ", " not like ", " =ilike ", " =like ", " not in ", " ilike ", " like ", " in ", " child_of ", "="}
	var splitString []string
	var boolInput bool
	result := []interface{}{}
	for i := 0; i < len(operators); i++ {
		if strings.Contains(strInput, operators[i]) {
			splitString = strings.Split(strInput, operators[i])

			// Trim trailing spaces in beginning and ending
			splitString[0] = strings.Trim(splitString[0], " ")
			splitString[1] = strings.Trim(splitString[1], " ")

			if splitString[len(splitString)-1] == "false" {
				boolInput = false
				result = []interface{}{splitString[0], strings.Trim(operators[i], " "), boolInput}
			} else if splitString[len(splitString)-1] == "true" {
				boolInput = true
				result = []interface{}{splitString[0], strings.Trim(operators[i], " "), boolInput}
			} else {
				result = []interface{}{splitString[0], strings.Trim(operators[i], " "), splitString[len(splitString)-1]}
			}
			return result, nil
		}
	}
	return nil, fmt.Errorf("wrong input in domain condition")
}
