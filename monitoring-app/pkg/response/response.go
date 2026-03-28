package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
    Status  string      `json:"status,omitempty"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, data interface{}) {
    JSON(w, http.StatusOK, Response{
        Status: "success",
        Data:   data,
    })
}

func Error(w http.ResponseWriter, status int, message string) {
    JSON(w, status, Response{
        Status: "error",
        Error:  message,
    })
}