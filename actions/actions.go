package actions

import (
	"encoding/json"
	"fmt"
	"go-auth/store"
	"net/http"
)

//HTTPAction is extended HttpHandler which returns http status of response and data
type HTTPAction func(*http.Request) (int, interface{})

//Run midleware for running actions action
func Run(action HTTPAction, method string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method != method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		status, data := action(r)

		if status == http.StatusInternalServerError {
			switch err := data.(type) {
			case error:
				internalError(w, err.Error())
				return
			}
		}

		response, err := json.Marshal(data)
		if err != nil {
			internalError(w, "Internal error while marshaling response")
			return
		}

		w.WriteHeader(status)
		w.Write(response)
	}
}

type healthCheckResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Healthcheck action for service
func Healthcheck(r *http.Request) (int, interface{}) {
	response := healthCheckResponse{"Ok", "I'm OK", http.StatusOK}
	return http.StatusOK, response
}

//Registration action for service
func Registration(r *http.Request) (int, interface{}) {
	var user store.User

	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&user); err != nil {
		return http.StatusBadRequest, nil
	}

	if ok, validationErrors, err := user.Create(); !ok {
		return http.StatusUnprocessableEntity, validationErrors
	} else if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func internalError(w http.ResponseWriter, msg string) {
	m := map[string]string{"error": msg}

	e, _ := json.Marshal(m)
	fmt.Println("Internal error:", msg)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(e))
}
