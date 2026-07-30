package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cocuum/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Event string `json:"event"`
		Data struct {
			UserID uuid.UUID `json:"user_id"`
		} `json:"data"`
	}

	key, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "no header included in request", err)
		return
	}

	if key != cfg.polka {
		respondWithError(w, http.StatusUnauthorized, "Invalid ApiKey", err)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not decode parameters", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = cfg.db.UpgradeToChirpRed(r.Context(), params.Data.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "could not find user", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "could upgrade user id", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
