package api

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/lite-dokploy/backend/internal/api/handlers"
	"github.com/lite-dokploy/backend/internal/api/middleware"
	"github.com/lite-dokploy/backend/internal/auth"
	"github.com/lite-dokploy/backend/internal/deploy"
)

func NewRouter(engine *deploy.Engine, frontendDir string, authManager *auth.Manager) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	dh := handlers.NewDeployHandler(engine)
	wh := handlers.NewWebhookHandler(engine)
	ah := handlers.NewAuthHandler(authManager)

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", ah.Login)
			r.Post("/logout", ah.Logout)
		})

		r.Post("/webhook/{id}", wh.HandleWebhook)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(authManager))

			r.Get("/auth/me", ah.Me)
			r.Get("/dashboard", handlers.GetDashboard)

			r.Route("/applications", func(r chi.Router) {
				r.Get("/", handlers.ListApps)
				r.Post("/", handlers.CreateApp)
				r.Get("/{id}", handlers.GetApp)
				r.Put("/{id}", handlers.UpdateApp)
				r.Delete("/{id}", handlers.DeleteApp)

				r.Post("/{id}/deploy", dh.DeployApp)
				r.Post("/{id}/redeploy", dh.RedeployApp)
				r.Post("/{id}/restart", dh.RestartApp)
				r.Post("/{id}/stop", dh.StopApp)
				r.Post("/{id}/start", dh.StartApp)

				r.Get("/{id}/deployments", dh.ListDeployments)
				r.Get("/{id}/logs", handlers.ListAppLogs)

				r.Route("/{id}/domains", func(r chi.Router) {
					r.Get("/", handlers.ListAppDomains)
					r.Post("/", handlers.CreateAppDomain)
					r.Put("/{domainId}", handlers.UpdateAppDomain)
					r.Delete("/{domainId}", handlers.DeleteAppDomain)
				})
			})

			r.Get("/deployments/{depId}/logs", handlers.GetDeploymentLog)
		})
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	if frontendDir != "" {
		if info, err := os.Stat(frontendDir); err == nil && info.IsDir() {
			fileServer := http.FileServer(http.Dir(frontendDir))
			fallback := func(w http.ResponseWriter, r *http.Request) {
				path := filepath.Join(frontendDir, r.URL.Path)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					http.ServeFile(w, r, filepath.Join(frontendDir, "index.html"))
					return
				}
				fileServer.ServeHTTP(w, r)
			}

			r.HandleFunc("/*", fallback)
		}
	}

	return r
}

type fsFunc func(name string) (fs.File, error)

func (f fsFunc) Open(name string) (fs.File, error) {
	return f(name)
}
