package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lite-dokploy/backend/internal/api"
	"github.com/lite-dokploy/backend/internal/database"
	"github.com/lite-dokploy/backend/internal/deploy"
	"github.com/lite-dokploy/backend/internal/docker"
	"github.com/lite-dokploy/backend/internal/git"
	"github.com/lite-dokploy/backend/internal/traefik"
)

func main() {
	port := flag.Int("port", 3000, "HTTP server port")
	dataDir := flag.String("data", "/data", "Data directory for SQLite DB")
	repoDir := flag.String("repos", "/repositories", "Directory for git repositories")
	logDir := flag.String("logs", "/logs", "Directory for deployment logs")
	frontendDir := flag.String("frontend", "", "Directory for frontend static files")
	traefikEnabled := flag.Bool("traefik", false, "Enable Traefik integration")
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting Lite Dokploy...")

	if err := database.Init(*dataDir); err != nil {
		log.Fatalf("Database init: %v", err)
	}
	defer database.Close()

	dockerManager, err := docker.NewManager()
	if err != nil {
		log.Fatalf("Docker init: %v", err)
	}
	defer dockerManager.Close()

	if err := dockerManager.Ping(context.Background()); err != nil {
		log.Printf("WARNING: Docker not available: %v", err)
	}

	gitManager := git.NewManager(*repoDir)
	traefikManager := traefik.NewManager()
	if *traefikEnabled {
		traefikManager.Enable()
	}

	deployEngine := deploy.NewEngine(dockerManager, gitManager, traefikManager)
	router := api.NewRouter(deployEngine, *frontendDir)

	for _, dir := range []string{*dataDir, *repoDir, *logDir, "/deployments"} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Create dir %s: %v", dir, err)
		}
	}

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Lite Dokploy listening on :%d", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced shutdown: %v", err)
	}

	log.Println("Server stopped")
}
