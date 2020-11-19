package api

import (
	"database/sql"
	"encoding/base64"
	"go-redis-pg/api/models"
	"go-redis-pg/api/utils"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if len(s) != 2 {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid/Missing Credentials.")
			return
		}

		b, err := base64.StdEncoding.DecodeString(s[1])
		if err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid/Missing Credentials.")
			return
		}

		pair := strings.SplitN(string(b), ":", 2)
		if len(pair) != 2 {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid/Missing Credentials.")
			return
		}

		user := models.User{Username: pair[0]}
		row := db.QueryRow("SELECT id, saltedpassword, salt FROM users WHERE username=?", user.Username)
		if err := row.Scan(&user.Id, &user.Saltedpassword, &user.Salt); err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid/Missing Credentials.")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.Saltedpassword), []byte(pair[1]+user.Salt)); err != nil {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid/Missing Credentials.")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) cacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method != "GET" {
			next.ServeHTTP(w, r)
			return
		}

		content, err := s.Cache.Get(r.RequestURI).Result()
		if err != nil {
			rr := httptest.NewRecorder()
			next.ServeHTTP(rr, r)
			content = rr.Body.String()
			err = s.Cache.Set(r.RequestURI, content, 10*time.Minute).Err()
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
			}
			utils.RespondWithString(w, http.StatusOK, content)
			return
		}
		utils.RespondWithString(w, http.StatusOK, content)
		return

	})
}
