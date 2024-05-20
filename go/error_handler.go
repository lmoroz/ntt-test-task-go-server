package swagger

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func CheckErrorFatal(err error, msg string) {
	if err != nil {
		log.Fatalf(msg+": %v", err)
	}
}

func checkErrorInternal(err error, w http.ResponseWriter) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}
	return false
}

func checkErrorBadRequest(err error, w http.ResponseWriter) bool {
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return true
	}
	return false
}

func checkSQLError(err error, w http.ResponseWriter, m string) bool {
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%s » Record not found", m)})
			return true
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("%s » %v", m, err.Error())})
		return true
	}
	return false
}
