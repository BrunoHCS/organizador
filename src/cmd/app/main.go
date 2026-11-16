package main

import (
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"errors"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

	"biblioteca-ssd/internal/api"
	"biblioteca-ssd/internal/config"
	"biblioteca-ssd/internal/db"
	"biblioteca-ssd/internal/scanner"
)

// Production builds place the Vite output in this directory.
//
//go:embed all:dist
var frontend embed.FS

func main() {
	if err := run(); err != nil {
		slog.Error("aplicação encerrada com erro", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	settings, err := config.Load()
	if err != nil {
		return err
	}
	logger.Info("raízes definidas", "app", settings.AppDir, "biblioteca", settings.LibraryRoot, "dados", settings.DataDir)

	catalog, err := db.Open(settings.DatabasePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := catalog.Close(); err != nil {
			logger.Error("falha ao fechar o catálogo", "error", err)
		}
	}()

	static, err := fs.Sub(frontend, "dist")
	if err != nil {
		return err
	}
	token, err := newSessionToken()
	if err != nil {
		return err
	}
	server := &http.Server{
		Handler:           api.NewCatalogHandler(static, token, catalog, scanner.Options{Root: settings.LibraryRoot, AppDir: settings.AppDir, Excluded: settings.ExcludedFolders}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	address := "http://" + listener.Addr().String()
	logger.Info("biblioteca iniciada", "url", address)
	openBrowser(address+"/?session="+token, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			logger.Error("falha ao encerrar o servidor", "error", err)
		}
	}()
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func newSessionToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func openBrowser(url string, logger *slog.Logger) {
	var command string
	var args []string
	switch runtime.GOOS {
	case "windows":
		command, args = "rundll32", []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		command, args = "open", []string{url}
	default:
		command, args = "xdg-open", []string{url}
	}
	if err := exec.Command(command, args...).Start(); err != nil {
		logger.Warn("não foi possível abrir o navegador automaticamente", "url", url, "error", err)
	}
}
