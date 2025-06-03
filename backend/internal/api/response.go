package api

import (
	"encoding/json"
	"net/http"

	"backend/internal/model"
)

// RespondWithError 返回错误响应
func RespondWithError(w http.ResponseWriter, code int, message string) {
	response := model.Response{
		Code:      code,
		Message:   message,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

// RespondWithJSON 返回JSON响应
func RespondWithJSON(w http.ResponseWriter, code int, message string, data any) {
	response := model.Response{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}
