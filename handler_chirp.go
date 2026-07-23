package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/cocuum/chirpy/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	ID			uuid.UUID `json:"id"`
	CreatedAt	time.Time `json:"created_at"`
	UpdatedAt	time.Time `json:"updated_at"`
	Body		string    `json:"body"`
	UserID		uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerCreateChirp(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body	string `json:"body"`
		UserID	uuid.UUID `json:"user_id"`
	}

	type responseBody struct {
		Chirp
	} 

	decoder := json.NewDecoder(r.Body)
	params := requestBody{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not decode parameters", err)
		return
	}

	clean_body, err := validateChirp(params.Body)
	if clean_body == "" {
		respondWithError(w, http.StatusBadRequest, err.Error(), err)
	}

	chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
		Body: clean_body,
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create chirp", err)
		return
	}

	respondWithJSON(w, http.StatusCreated, responseBody{
		Chirp: Chirp{
			ID:			chirp.ID,
			CreatedAt:	chirp.CreatedAt,
			UpdatedAt:	chirp.UpdatedAt,
			Body:		chirp.Body,
			UserID:		chirp.UserID,
		},
	})
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not get chirps", err)
		return
	}
	
	var allChirps = []Chirp{}
	for _, chirp := range dbChirps {
		
		allChirps = append(allChirps, Chirp{
			ID: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserID: chirp.UserID,
		})
	}

	respondWithJSON(w, http.StatusOK, allChirps)
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	type responseBody struct {
		Chirp
	} 

	dbID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not parse ID", err)
		return
	}

	dbChirp , err := cfg.db.GetChirpByID(r.Context(), dbID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, responseBody{
		Chirp: Chirp{
			ID:			dbChirp.ID,
			CreatedAt:	dbChirp.CreatedAt,
			UpdatedAt:	dbChirp.UpdatedAt,
			Body:		dbChirp.Body,
			UserID:		dbChirp.UserID,
		},
	})
}

func validateChirp(body string) (string,error) {
	const maxChirpLength = 140
	if len(body) > maxChirpLength {
		return "", errors.New("Chirp is too long")
	}

	badWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	cleaned := cleanBody(body, badWords)
	return cleaned,nil
}

func cleanBody(body string, badWords map[string]struct{}) string {
	words := strings.Split(body, " ")
	for i,word := range words{
		loweredWord := strings.ToLower(word)
		if _,ok := badWords[loweredWord]; ok {
			words[i] = "****"
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}