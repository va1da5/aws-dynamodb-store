package server

import (
	"aws-dynamodb-store/internal/domain"
	"aws-dynamodb-store/internal/model"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger) // Chi's built-in logger
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(s.AuthMiddleware)

	r.Get("/", s.HelloWorldHandler)
	r.Post("/users", s.CreateUser)
	r.Get("/users", s.GetUsers)
	r.Get("/{userID}", s.GetUser)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	return r
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	// For RBAC: Check if the current actor (e.g., from JWT) has "user:create" permission
	// actorUserID, _ := r.Context().Value("userID").(string) // Assuming userID is in context after auth
	// hasPerm, err := h.rbacService.HasPermission(r.Context(), actorUserID, "user:create")
	// if err != nil || !hasPerm {
	// 	http.Error(w, "Forbidden", http.StatusForbidden)
	// 	return
	// }

	var user model.UserCreateInput
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		writeJSONError(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// You'd typically generate a unique ID for the user here or in the service
	// user.EntityID = ...

	createdUser, err := s.service.RBACService.CreateUser(r.Context(), user.DisplayName, user.Email)
	if err != nil {
		log.Print(err)
		writeJSONError(w, "Failed to create user", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(createdUser)
}

// GetUser handles GET /users/{userID}
func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		writeJSONError(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Optional RBAC check: Does the current actor have permission to view this specific user?
	// Or, is the actor the user themselves?
	// actorUserID, _ := r.Context().Value("userID").(string)
	// if actorUserID != userID {
	//  	hasPerm, err := h.rbacService.HasPermission(r.Context(), actorUserID, "user:read:"+userID) // Example specific perm
	//  	if err != nil || !hasPerm {
	//  		http.Error(w, "Forbidden", http.StatusForbidden)
	//  		return
	//  	}
	// }

	user, err := s.repository.User.GetUserByID(r.Context(), domain.UserID(userID))
	if err != nil {
		writeJSONError(w, "User ID is required", http.StatusBadRequest)
		return
	}
	if user == nil {
		writeJSONError(w, "User not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *Server) GetUsers(w http.ResponseWriter, r *http.Request) {

	users, err := s.repository.User.ListAllUsers(r.Context())
	if err != nil {
		log.Println(err.Error())
		writeJSONError(w, "Failed to get users", http.StatusInternalServerError)
		return
	}
	if users == nil {
		writeJSONError(w, "Users not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
