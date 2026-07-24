package main

import (
	"encoding/json"
	"net/http"

	"github.com/cocuum/chirpy/internal/auth"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	type responseBody struct {
		User
	} 

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not decode parameters", err)
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "could not find user", err)
		return
	}

	match, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not check phash, man!", err)
		return
	}

	if match {
		respondWithJSON(w, http.StatusOK, responseBody{
			User: User{
				ID: user.ID,
				CreatedAt: user.CreatedAt,
				UpdatedAt: user.UpdatedAt,
				Email: user.Email,
			},
		})
	} else {
		respondWithError(w, http.StatusUnauthorized, "Password don't match man!", err)
		return
	}
}
