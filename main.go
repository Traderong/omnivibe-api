package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Traderong/omnivibe-api/auth"
	"github.com/Traderong/omnivibe-api/config"
	"github.com/Traderong/omnivibe-api/database"
	"github.com/Traderong/omnivibe-api/handlers"
	"github.com/Traderong/omnivibe-api/middleware"
	"github.com/Traderong/omnivibe-api/migrations"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env when available.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; using system environment variables")
	}

	// Validate application configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Configuration error:", err)
	}

	_ = cfg

	log.Println("Configuration validated successfully")

	// Connect to PostgreSQL.
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Run database migrations.
	migrationCtx, migrationCancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer migrationCancel()

	if err := migrations.Run(migrationCtx, db); err != nil {
		log.Fatal("Database migrations failed:", err)
	}

	log.Println("Database migrations completed")

	// Refresh-token cleanup.
	cleanupRefreshTokens := func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cleanupCancel()

		if err := auth.CleanupRefreshTokens(cleanupCtx); err != nil {
			log.Printf("refresh token cleanup failed: %v", err)
			return
		}

		log.Println("refresh token cleanup completed")
	}

	cleanupRefreshTokens()

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			cleanupRefreshTokens()
		}
	}()

	// Authentication rate limiter.
	authRateLimiter := middleware.NewRateLimiter(10, time.Minute)

	go func() {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			authRateLimiter.Cleanup(30 * time.Minute)
		}
	}()

	// ------------------------------------------------------------
	// Authentication endpoints
	// ------------------------------------------------------------

	http.Handle(
		"/api/auth/register",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.Register),
		),
	)

	http.Handle(
		"/api/auth/login",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.Login),
		),
	)

	http.Handle(
		"/api/auth/verify-email",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.VerifyEmail),
		),
	)

	http.Handle(
		"/api/auth/resend-verification",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.ResendVerification),
		),
	)

	http.Handle(
		"/api/auth/forgot-password",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.ForgotPassword),
		),
	)

	http.Handle(
		"/api/auth/reset-password",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.ResetPassword),
		),
	)

	http.Handle(
		"/api/auth/refresh",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.Refresh),
		),
	)

	http.Handle(
		"/api/auth/logout",
		authRateLimiter.Middleware(
			http.HandlerFunc(handlers.Logout),
		),
	)

	// ------------------------------------------------------------
	// Authenticated account endpoints
	// ------------------------------------------------------------

	http.Handle(
		"/api/me",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.Me),
		),
	)

	http.Handle(
		"/api/me/profile",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.UpdateProfile),
		),
	)

	http.Handle(
		"/api/me/password",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.ChangePassword),
		),
	)

	http.Handle(
		"/api/me/comments",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.GetUserComments),
		),
	)

	// ------------------------------------------------------------
	// Notification routes
	// ------------------------------------------------------------

	// GET /api/notifications
	http.Handle(
		"/api/notifications",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.GetNotifications),
		),
	)

	// Routes under /api/notifications/{...}
	http.HandleFunc(
		"/api/notifications/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(
				r.URL.Path,
				"/api/notifications/",
			)

			// GET /api/notifications/unread-count
			if path == "unread-count" {
				auth.AuthMiddleware(
					http.HandlerFunc(handlers.GetUnreadNotificationCount),
				).ServeHTTP(w, r)
				return
			}

			// POST /api/notifications/read-all
			if path == "read-all" {
				auth.AuthMiddleware(
					http.HandlerFunc(handlers.MarkAllNotificationsRead),
				).ServeHTTP(w, r)
				return
			}

			// PATCH /api/notifications/{id}/read
			if strings.HasSuffix(path, "/read") {
				auth.AuthMiddleware(
					http.HandlerFunc(handlers.MarkNotificationRead),
				).ServeHTTP(w, r)
				return
			}

			http.Error(
				w,
				"not found",
				http.StatusNotFound,
			)
		},
	)

	// ------------------------------------------------------------
	// Post routes
	// ------------------------------------------------------------

	// POST /api/posts
	http.Handle(
		"/api/posts",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.CreatePost),
		),
	)

	// All routes under /api/posts/{...}
	http.HandleFunc(
		"/api/posts/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(
				r.URL.Path,
				"/api/posts/",
			)

			// ----------------------------------------------------
			// Comments
			// POST /api/posts/{id}/comments
			// GET  /api/posts/{id}/comments
			// ----------------------------------------------------
			if strings.HasSuffix(path, "/comments") {
				switch r.Method {
				case http.MethodPost:
					auth.AuthMiddleware(
						http.HandlerFunc(handlers.CreateComment),
					).ServeHTTP(w, r)

				case http.MethodGet:
					handlers.GetPostComments(w, r)

				default:
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
				}

				return
			}

			// ----------------------------------------------------
			// Engagement
			// GET /api/posts/{id}/engagement
			// ----------------------------------------------------
			if strings.HasSuffix(path, "/engagement") {
				if r.Method != http.MethodGet {
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
					return
				}

				auth.AuthMiddleware(
					http.HandlerFunc(handlers.PostEngagement),
				).ServeHTTP(w, r)

				return
			}

			// ----------------------------------------------------
			// Likes
			// POST   /api/posts/{id}/like
			// DELETE /api/posts/{id}/like
			// ----------------------------------------------------
			if strings.HasSuffix(path, "/like") {
				switch r.Method {
				case http.MethodPost:
					auth.AuthMiddleware(
						http.HandlerFunc(handlers.LikePost),
					).ServeHTTP(w, r)

				case http.MethodDelete:
					auth.AuthMiddleware(
						http.HandlerFunc(handlers.UnlikePost),
					).ServeHTTP(w, r)

				default:
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
				}

				return
			}

			// ----------------------------------------------------
			// Favorites
			// POST   /api/posts/{id}/favorite
			// DELETE /api/posts/{id}/favorite
			// ----------------------------------------------------
			if strings.HasSuffix(path, "/favorite") {
				switch r.Method {
				case http.MethodPost:
					auth.AuthMiddleware(
						http.HandlerFunc(handlers.FavoritePost),
					).ServeHTTP(w, r)

				case http.MethodDelete:
					auth.AuthMiddleware(
						http.HandlerFunc(handlers.UnfavoritePost),
					).ServeHTTP(w, r)

				default:
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
				}

				return
			}

			// ----------------------------------------------------
			// Delete post
			// DELETE /api/posts/{id}
			// ----------------------------------------------------
			if r.Method == http.MethodDelete {
				auth.AuthMiddleware(
					http.HandlerFunc(handlers.DeletePost),
				).ServeHTTP(w, r)

				return
			}

			// ----------------------------------------------------
			// Get post
			// GET /api/posts/{id}
			// ----------------------------------------------------
			if r.Method == http.MethodGet {
				handlers.GetPost(w, r)
				return
			}

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		},
	)

	// ------------------------------------------------------------
	// Comment routes outside a post
	// ------------------------------------------------------------

	// DELETE /api/comments/{id}
	http.Handle(
		"/api/comments/",
		auth.AuthMiddleware(
			http.HandlerFunc(handlers.DeleteComment),
		),
	)

	// ------------------------------------------------------------
	// User/profile/follow routes
	// ------------------------------------------------------------

	http.HandleFunc(
		"/api/users/",
		func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimSuffix(
				r.URL.Path,
				"/",
			)

			// Follow status:
			// GET /api/users/{username}/follow-status
			if strings.HasSuffix(path, "/follow-status") {
				if r.Method != http.MethodGet {
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
					return
				}

				auth.AuthMiddleware(
					http.HandlerFunc(handlers.FollowStatus),
				).ServeHTTP(w, r)

				return
			}

			// Follow/unfollow:
			// POST   /api/users/{username}/follow
			// DELETE /api/users/{username}/follow
			if strings.HasSuffix(path, "/follow") {
				auth.AuthMiddleware(
					http.HandlerFunc(
						func(w http.ResponseWriter, r *http.Request) {
							switch r.Method {
							case http.MethodPost:
								handlers.FollowUser(w, r)

							case http.MethodDelete:
								handlers.UnfollowUser(w, r)

							default:
								http.Error(
									w,
									"method not allowed",
									http.StatusMethodNotAllowed,
								)
							}
						},
					),
				).ServeHTTP(w, r)

				return
			}

			// Profile statistics:
			// GET /api/users/{username}/stats
			if strings.HasSuffix(path, "/stats") {
				if r.Method != http.MethodGet {
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
					return
				}

				handlers.FollowStats(w, r)
				return
			}

			// Followers:
			// GET /api/users/{username}/followers
			if strings.HasSuffix(path, "/followers") {
				if r.Method != http.MethodGet {
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
					return
				}

				handlers.Followers(w, r)
				return
			}

			// Following:
			// GET /api/users/{username}/following
			if strings.HasSuffix(path, "/following") {
				if r.Method != http.MethodGet {
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
					return
				}

				handlers.Following(w, r)
				return
			}

			// User posts:
			// GET /api/users/{username}/posts
			if strings.HasSuffix(path, "/posts") {
				if r.Method != http.MethodGet {
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
					return
				}

				handlers.GetUserPosts(w, r)
				return
			}

			// Public profile:
			// GET /api/users/{username}
			if r.Method != http.MethodGet {
				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
				return
			}

			handlers.PublicProfile(w, r)
		},
	)

	// ------------------------------------------------------------
	// Global request-body limit
	// ------------------------------------------------------------

	rootHandler := middleware.BodyLimit(
		http.DefaultServeMux,
	)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           rootHandler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    32 * 1024,
	}

	log.Println(
		"OmniVibe backend started on http://localhost:8080",
	)
	log.Println(
		"PostgreSQL connected successfully",
	)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal("HTTP server failed:", err)
	}
}
