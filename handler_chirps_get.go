package main

import (
	"errors"
	"net/http"
	"sort"

	"github.com/cocuum/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	type responseBody struct {
		Chirp
	}

	dbID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not parse ID", err)
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(r.Context(), dbID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Chirp not found", err)
		return
	}

	respondWithJSON(w, http.StatusOK, responseBody{
		Chirp: Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			Body:      dbChirp.Body,
			UserID:    dbChirp.UserID,
		},
	})
}

func authorID(r *http.Request) (uuid.UUID, error) {
	s := r.URL.Query().Get("author_id")
	if s == "" {
		return uuid.Nil, nil
	}
	authorID, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, err
	}
	return authorID, nil
}

func sortChirps(r *http.Request) (string, error) {
	s := r.URL.Query().Get("sort")
	switch s {
	case "" :
		return "asc",nil
	case "asc":
		return "asc", nil
	case "desc":
		return "desc", nil
	default:
		return "", errors.New("Invalid Sort")
	}
}

func prepareResponsePayload(db []database.Chirp) []Chirp {
	var allChirps = []Chirp{}
	for _, chirp := range db {
			allChirps = append(allChirps, Chirp{
				ID:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserID:    chirp.UserID,
			})
		}
	return allChirps
}

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	authorID, err := authorID(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid author ID", err)
	}

	var dbChirps []database.Chirp
	
	if authorID != uuid.Nil {
		dbChirps, err = cfg.db.GetChirpByUserID(r.Context(), authorID)
	} else {
		dbChirps, err = cfg.db.GetAllChirps(r.Context())
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not retrieve chirps", err)
	}

	sortDirection, err := sortChirps(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid sort value", err)
	}

	chirps := prepareResponsePayload(dbChirps)

	sort.Slice(chirps, func(i, j int) bool {
		if sortDirection == "desc" {
			return chirps[i].CreatedAt.After(chirps[j].CreatedAt)
		}
		return chirps[i].CreatedAt.Before(chirps[j].CreatedAt)
	})

	respondWithJSON(w, http.StatusOK, chirps)
}
