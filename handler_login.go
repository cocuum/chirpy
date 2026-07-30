package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cocuum/chirpy/internal/auth"
	"github.com/cocuum/chirpy/internal/database"
)



func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Password string `json:"password"`
		Email string `json:"email"`
	}

	type responseBody struct {
		User
		Token string `json:"token"`
		RefreshToken	string `json:"refresh_token"`
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
	if err != nil || !match{
		respondWithError(w, http.StatusUnauthorized, "Password don't match man!", err)
		return
	}

	token, err := auth.MakeJWT(
		user.ID,
		cfg.jwtsecret,
		time.Hour,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not make access token", err)
		return
	}

	refreshToken := auth.MakeRefreshToken()
	_, err = cfg.db.CreateRefreshTokens(r.Context(), database.CreateRefreshTokensParams{
		UserID: user.ID,
		Token: refreshToken,
		ExpiresAt: time.Now().Add(24*60*time.Hour),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create refresh token", err)
		return
	}


	respondWithJSON(w, http.StatusOK, responseBody{
		User: User{
			ID:				user.ID,
			CreatedAt:		user.CreatedAt,
			UpdatedAt:		user.UpdatedAt,
			Email:			user.Email,
			IsChirpyRed:	user.IsChirpyRed,
		},
		Token: token,
		RefreshToken: refreshToken,
	})
}
