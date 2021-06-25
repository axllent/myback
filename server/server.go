package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/axllent/myback/logger"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	_ "github.com/go-sql-driver/mysql"
)

var limitToUsers = make(map[string]bool)
var limitToIPs = make(map[string]bool)

func Listen() error {
	checkConfig()

	proto := "http"
	if Config.SSLCert != "" && Config.SSLKey != "" {
		proto = "https"
	}

	serverAddr := fmt.Sprintf("%s://%s", proto, Config.Listen)
	dbAddr := fmt.Sprintf("%s:%d", Config.MySQLHost, Config.MySQLPort)
	logger.Log().Noticef("Starting server on %s for database %s", serverAddr, dbAddr)

	r := mux.NewRouter().StrictSlash(true)

	r.HandleFunc("/db", basicAuth(listDatabases))
	r.HandleFunc("/db/{database}", basicAuth(listTables))
	r.HandleFunc("/dump", basicAuth(dump))
	r.HandleFunc("/dump/{database}", basicAuth(dump))
	r.HandleFunc("/dump/{database}/{table}", basicAuth(dump))
	r.HandleFunc("/health", health)
	r.NotFoundHandler = basicAuth(pageNotFound)

	http.Handle("/", r)

	gzip := handlers.CompressHandler(r)

	if Config.SSLCert != "" && Config.SSLKey != "" {
		return http.ListenAndServeTLS(Config.Listen, Config.SSLCert, Config.SSLKey, gzip)
	}

	return http.ListenAndServe(Config.Listen, gzip)
}

// BasicAuth uses MySQL login details
func basicAuth(handler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		user, pass, ok := r.BasicAuth()

		if !ok {
			basicAuthResponse(w)
			return
		}

		if len(limitToIPs) > 0 {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				logger.Log().Error(err.Error())
				basicAuthResponse(w)
				return
			}
			_, ok = limitToIPs[ip]
			if !ok {
				logger.Log().Errorf("Unauthorised IP: %s", ip)
				basicAuthResponse(w)
				return
			}
		}

		// ensure only limited users (if set) can access
		_, ok = limitToUsers[user]
		if len(limitToUsers) > 0 && !ok {
			basicAuthResponse(w)
			return
		}

		// open the connection
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", user, pass, Config.MySQLHost, Config.MySQLPort))
		if err != nil {
			basicAuthResponse(w)
			return
		}
		defer db.Close()

		// ensure we can connect
		err = db.Ping()
		if err != nil {
			basicAuthResponse(w)
			return
		}

		handler(w, r)
	}
}

// JSONResponse returns a 200 JSON response to the browser
func jsonResponse(w http.ResponseWriter, obj interface{}) {
	results, err := json.Marshal(obj)
	if err != nil {
		httpError(w, err)
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "%s", string(results))
}

// BasicAuthResponse returns an basic auth response to the browser
func basicAuthResponse(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="Login"`)
	w.WriteHeader(http.StatusUnauthorized)
	if _, err := w.Write([]byte("Unauthorised.\n")); err != nil {
		logger.Log().Error(err.Error())
	}
}

// HTTPError returns an error response to the browser
func httpError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// PageNotFound custom handler so we can intercept
// everything with basic auth
func pageNotFound(w http.ResponseWriter, r *http.Request) {
	logger.Log().Errorf("Not found: %s %s", ip(r), r.URL.Path)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "Page not found")
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
