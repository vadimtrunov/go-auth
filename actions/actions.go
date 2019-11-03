package actions

import (
	"encoding/json"
	"go-auth/store"
	"net/http"

	"github.com/asaskevich/govalidator"
)

//HTTPAction is extended HttpHandler which returns http status of response and data
type HTTPAction func(*http.Request) (int, interface{})

//Run midleware for running actions action
func Run(action HTTPAction, method string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.WriteHeader(405)
			return
		}

		status, data := action(r)

		response, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal error while marshaling response"))
			return
		}
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
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

	if ok, err := govalidator.ValidateStruct(user); !ok {
		return http.StatusUnprocessableEntity, govalidator.ErrorsByField(err)
	} else if err != nil {
		return http.StatusInternalServerError, nil
	}

	if err := user.Save(); err != nil {
		return http.StatusInternalServerError, nil
	}

	return http.StatusNoContent, nil
}
