package main

import (
	"encoding/json"
	"net/http"

	"github.com/cocuum/chirpy/internal/auth"
	"github.com/cocuum/chirpy/internal/database"
)

func (cfg *apiConfig) handlerUpdateUser(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	type responseBody struct {
		User
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not find jwt", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtsecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "could not validate jwt", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not decode parameters", err)
		return
	}

	hp, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not hash this pword, man!", err)
		return
	}

	user, err := cfg.db.UpdateEmailHashByUser(
		r.Context(),
		database.UpdateEmailHashByUserParams{
			Email:          params.Email,
			HashedPassword: hp,
			ID:             userID,
		},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not update user", err)
	}

	respondWithJSON(w, http.StatusOK, responseBody{
		User: User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email:     user.Email,
		},
	})
}
