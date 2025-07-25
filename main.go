package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Osirous/A_Clicker_Game_server/internal/auth"
	"github.com/Osirous/A_Clicker_Game_server/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiHandler struct{}

type apiConfig struct {
	DB        *database.Queries
	jwtSecret string
	jwtIssuer string
}

type User struct {
	ID             uuid.UUID `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Username       string    `json:"username"`
	HashedPassword string    `json:"-"` // - means do not include this field in client JSON responses.
	Token          string    `json:"token"`
	RefreshToken   string    `json:"refresh_token"`
	SaveID         uuid.UUID `json:"save_id"`
}

type SaveDataRequest struct {
	Savedata string `json:"savedata"`
}

type Save struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Savedata  string    `json:"savedata"`
	UserID    uuid.UUID `json:"user_id"`
}

func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func writeJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (cfg *apiConfig) saveHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:

		saveIDStr := r.PathValue("saveID")
		saveID, err := uuid.Parse(saveIDStr)
		if err != nil {
			writeJSONError(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		save, err := cfg.DB.GetSaveData(r.Context(), saveID)
		if err != nil {
			writeJSONError(w, "Save not found!", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(save.Savedata))

	case http.MethodPut:

		saveIDStr := r.PathValue("saveID")
		saveID, err := uuid.Parse(saveIDStr)
		if err != nil {
			writeJSONError(w, "Invalid UUID format", http.StatusBadRequest)
			return
		}

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			log.Printf("Unauthorized access: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.jwtSecret, cfg.jwtIssuer)
		if err != nil {
			log.Printf("Unauthorized access: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			writeJSONError(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		var req SaveDataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Savedata == "" {
			writeJSONError(w, "Missing savedata", http.StatusBadRequest)
			return
		}

		saveParams := database.UpdateSaveDataParams{
			ID:       saveID,
			Savedata: req.Savedata,
			UserID:   userID,
		}

		_, err = cfg.DB.UpdateSaveData(r.Context(), saveParams)
		if err != nil {
			writeJSONError(w, "Failed to save user data", http.StatusInternalServerError)
			return
		}

		updatedSave, err := cfg.DB.GetSaveData(r.Context(), saveID)
		if err != nil {
			writeJSONError(w, "Failed to get user data", http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedSave)

	case http.MethodPost:

		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			log.Printf("Unauthorized access: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := auth.ValidateJWT(token, cfg.jwtSecret, cfg.jwtIssuer)
		if err != nil {
			log.Printf("Unauthorized access: %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			writeJSONError(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}

		// Decode JSON input to get base64-encoded savedata
		var req SaveDataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		if req.Savedata == "" {
			writeJSONError(w, "Missing savedata", http.StatusBadRequest)
			return
		}

		saveParams := database.CreateSaveDataParams{
			Savedata: req.Savedata,
			UserID:   userID,
		}
		save, err := cfg.DB.CreateSaveData(r.Context(), saveParams)
		if err != nil {
			writeJSONError(w, "Failed to save user data", http.StatusInternalServerError)
			return
		}

		newSave := Save{
			ID:        save.ID,
			CreatedAt: save.CreatedAt,
			UpdatedAt: save.UpdatedAt,
			UserID:    userID,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newSave)

	default:
		writeJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
}

func (cfg *apiConfig) userHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodPost:

		type parameters struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			writeJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		hashedPassword, err := auth.HashPassword(params.Password)
		if err != nil {
			writeJSONError(w, "Error processing password", http.StatusInternalServerError)
			return
		}

		userParams := database.CreateUserParams{
			Username:       params.Username,
			HashedPassword: hashedPassword,
		}

		user, err := cfg.DB.CreateUser(r.Context(), userParams)
		if err != nil {
			writeJSONError(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		responseUser := User{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Username:  user.Username,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(responseUser)

	default:
		writeJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodPost:

		type parameters struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}

		decoder := json.NewDecoder(r.Body)
		params := parameters{}
		err := decoder.Decode(&params)
		if err != nil {
			writeJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		getUser, err := cfg.DB.GetUserByUsername(r.Context(), params.Username)
		if err != nil {
			writeJSONError(w, "Incorrect username or password", http.StatusUnauthorized)
			return
		}

		err = auth.CheckPasswordHash(getUser.HashedPassword, params.Password)
		if err != nil {
			writeJSONError(w, "Incorrect username or password", http.StatusUnauthorized)
			return
		}

		expiresIn := time.Hour // default 1 hour

		accessToken, err := auth.MakeJWT(getUser.ID, cfg.jwtSecret, cfg.jwtIssuer, expiresIn)
		if err != nil {
			writeJSONError(w, "Error creating token", http.StatusInternalServerError)
			return
		}

		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			writeJSONError(w, "Error creating refresh token", http.StatusInternalServerError)
			return
		}

		expiresAt := time.Now().AddDate(0, 0, 60) // 60 days from now

		rtParams := database.StoreRefreshTokenParams{
			Token:     refreshToken,
			UserID:    uuid.NullUUID{UUID: getUser.ID, Valid: true},
			ExpiresAt: sql.NullTime{Time: expiresAt, Valid: true},
		}

		storedRefreshToken, err := cfg.DB.StoreRefreshToken(r.Context(), rtParams)
		if err != nil {
			writeJSONError(w, "Error storing refresh token", http.StatusInternalServerError)
			return
		}

		getSaveID, err := cfg.DB.GetSaveDataByUserID(r.Context(), getUser.ID)
		var saveID uuid.UUID

		if err != nil {
			if err == sql.ErrNoRows {
				// User has no save, that's okay! They are new!
				// Set saveID to default zero value
				saveID = uuid.Nil
			} else {
				writeJSONError(w, "Error getting save id", http.StatusInternalServerError)
				return
			}
		} else {
			saveID = getSaveID.ID
		}

		response := User{
			ID:           getUser.ID,
			CreatedAt:    getUser.CreatedAt,
			UpdatedAt:    getUser.UpdatedAt,
			Username:     getUser.Username,
			Token:        accessToken,
			RefreshToken: storedRefreshToken.Token,
			SaveID:       saveID,
		}

		data, err := json.Marshal(response)
		if err != nil {
			writeJSONError(w, "Error encoding response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)

	default:
		writeJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func main() {

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	// sql db stuff
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	// api config
	apiCfg := &apiConfig{
		DB:        dbQueries,
		jwtSecret: os.Getenv("JWT_SECRET"),
		jwtIssuer: os.Getenv("JWT_ISSUER"),
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", apiHandler{})
	mux.HandleFunc("/api/users", apiCfg.userHandler)
	mux.HandleFunc("/api/login", apiCfg.loginHandler)
	mux.HandleFunc("/api/refresh", apiCfg.refreshHandler)
	mux.HandleFunc("/api/revoke", apiCfg.revokedHandler)
	mux.HandleFunc("/api/savedata/{saveID}", apiCfg.saveHandler) // Handle POST (Create)
	mux.HandleFunc("/api/savedata", apiCfg.saveHandler)          // Handle GET, PUT (by id)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		// Error starting or closing listener:
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}

}

func (cfg *apiConfig) revokedHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		// Extract refresh token from request header
		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			log.Printf("Unauthorized refresh access: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Revoke token here

		_, err = cfg.DB.RevokeToken(r.Context(), refreshToken)
		if err != nil {
			writeJSONError(w, "Unabled to revoke token", http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusNoContent)

	default:
		writeJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		// Extract refresh token from request header
		refreshToken, err := auth.GetBearerToken(r.Header)
		if err != nil {
			log.Printf("Could not find token: %s", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Check that the token exists in your data store
		user, err := cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken)
		if err != nil {
			writeJSONError(w, "Invalid or missing refresh token", http.StatusUnauthorized)
			return
		}

		// Issue a new access token
		newAccessToken, err := auth.MakeJWT(user.ID, cfg.jwtSecret, cfg.jwtIssuer, time.Hour)
		if err != nil {
			writeJSONError(w, "Error creating new token", http.StatusInternalServerError)
			return
		}

		// Respond with the new access token
		response := struct {
			Token string `json:"token"`
		}{
			Token: newAccessToken,
		}

		data, err := json.Marshal(response)
		if err != nil {
			writeJSONError(w, "Error encoding response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)

	default:
		writeJSONError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

}
