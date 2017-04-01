package main

import (
	"encoding/json"
	"net/http"

	"base"
)

func (a *App) IndexHandler(db *base.DB) HandlerWithError {
	return func(w http.ResponseWriter, req *http.Request) error {
		result := struct {
			Status string `json:"status"`
		}{
			Status: "success",
		}
		err := json.NewEncoder(w).Encode(result)
		if err != nil {
			return newAPIError(500, "error when encoding response, %s", err)
		}

		return nil
	}
}
