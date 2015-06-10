package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func writeJSON(w http.ResponseWriter, v interface{}, code int) {

	jsonData, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	w.Write(jsonData)

}

func getIDFromMux(vars map[string]string) (id int, err error) {
	ids, ok := vars["id"]
	if !ok {
		return 0, errors.New("no ID supplied")
	}

	id, err = strconv.Atoi(ids)
	if err != nil {
		return 0, err
	}
	return id, nil
}
