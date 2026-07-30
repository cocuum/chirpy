package main

import (
	"net/http"

	"github.com/cocuum/chirpy/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {

	dbID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not parse ID", err)
		return
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

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), dbID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	if dbChirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "Invalid user", err)
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), dbID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not del chirp", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
