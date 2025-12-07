package main

import (
	"auth/internal/delivery/http/middleware"
	"auth/internal/delivery/http/response"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	openApiMiddleware "github.com/go-openapi/runtime/middleware"
)

func (a *App) setupRouter(handler *Handler) chi.Router {
	router := chi.NewRouter()

	// Global middleware
	router.Use(chiMiddleware.Logger)
	router.Use(chiMiddleware.Recoverer)

	router.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		response.WriteError(w, http.StatusNotFound, errors.New("method is not allowed"), a.logger)
	})

	router.NotFound(func(w http.ResponseWriter, r *http.Request) {
		response.WriteError(w, http.StatusNotFound, errors.New("route doesn't exists"), a.logger)
	})

	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		response.WriteSuccess(w, http.StatusOK, map[string]string{"message": "server is healthy"})
	})

	// API v1 routes
	setupAPIRoutes(router, handler)

	return router
}

// serveOpenAPISpec serves the OpenAPI specification file
func serveOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/x-yaml")
	http.ServeFile(w, r, openAPIPath)
}

// setupSwaggerUI configures Swagger UI routes
func setupSwaggerUI(router chi.Router) {
	opts := openApiMiddleware.SwaggerUIOpts{
		SpecURL: "/api/v1/openapi.yaml",
		Path:    "/api/v1/docs",
	}
	sh := openApiMiddleware.SwaggerUI(opts, nil)
	router.Handle("/docs", sh)
	router.Handle("/docs/*", sh)
}

func setupAPIRoutes(router chi.Router, handler *Handler) {
	router.Route("/api/v1", func(api chi.Router) {
		// OpenAPI specification (inside /api/v1)
		api.Get("/openapi.yaml", serveOpenAPISpec)

		// Swagger UI documentation (inside /api/v1)
		setupSwaggerUI(api)

		// Utilities / test routes
		api.Get("/geo", handler.test.GetGeoLocation)

		// Auth routes
		authRoutes(api, handler)

		// User routes
		userRoutes(api, handler)
	})
}

func authRoutes(api chi.Router, handler *Handler) {
	api.Route("/auth", func(auth chi.Router) {
		auth.Post("/", handler.auth.LoginUser)
		auth.Post("/refresh", handler.auth.RefreshToken)
		auth.Get("/verify-email", handler.auth.VerifyEmail)
		auth.Post("/request-code", handler.auth.RequestVerificationCode)
		auth.Post("/verify-code", handler.auth.VerifyCode)
		auth.Post("/password/request-reset", handler.auth.RequestPasswordReset)
		auth.Post("/password/reset", handler.auth.ResetPassword)
		auth.Post("/otp/request", handler.auth.RequestLoginOTP)
		auth.Post("/otp/verify", handler.auth.VerifyLoginOTP)
		auth.Post("/logout", handler.auth.Logout)
	})
}

func userRoutes(api chi.Router, handler *Handler) {
	api.Route("/users", func(user chi.Router) {
		user.Post("/", handler.user.RegisterUserWithCode)

		// Protected route
		user.Group(func(user chi.Router) {
			user.Use(middleware.AuthMiddleware)
			user.Get("/me", handler.user.GetUserProfile)
		})
	})
}
