package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-students/app/repository"
	"api-students/app/service"
	"api-students/config"
	"api-students/database"
	"api-students/helper"
	"api-students/route"
)

func main() {
	config.LoadEnv()
	logger := config.NewLogger()

	jwtSecret := config.GetEnv("JWT_SECRET", "super-secret-key-that-is-at-least-32-chars-long")
	if len(jwtSecret) < 32 {
		logger.Error("JWT_SECRET minimal 32 karakter")
		os.Exit(1)
	}

	pool, err := database.NewPool(context.Background())
	if err != nil {
		logger.Error("gagal terhubung ke database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	jwtManager := helper.NewJWTManager(
		jwtSecret,
		config.GetEnv("JWT_ISSUER", "praktikum-backend"),
		time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 15))*time.Minute,
	)

	// Inisialisasi Repository
	studentRepo := repository.NewStudentRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	tokenRepo := repository.NewTokenRepository(pool)
	roleRepo := repository.NewRoleRepository(pool)

	// Muat RBAC Permissions saat startup
	rawPermissions, err := roleRepo.LoadPermissions(context.Background())
	if err != nil {
		logger.Error("gagal memuat permission", slog.String("error", err.Error()))
		os.Exit(1)
	}
	perms := helper.NewPermissionSet(rawPermissions)
	logger.Info("permission RBAC dimuat", slog.Any("roles", perms.KnownRoles()))

	// Inisialisasi Service
	studentService := service.NewStudentService(studentRepo, perms)
	authService := service.NewAuthService(
		userRepo, tokenRepo, jwtManager, perms,
		time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7))*24*time.Hour,
	)

	// Inisialisasi Web App & Router
	app := config.NewApp(logger)
	route.Register(app, route.Dependencies{
		Pool:           pool,
		JWT:            jwtManager,
		Permissions:    perms,
		StudentService: studentService,
		AuthService:    authService,
	})

	port := config.GetEnv("APP_PORT", "3000")
	go func() {
		if err := app.Listen(":" + port); err != nil {
			logger.Error("server berhenti", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	logger.Info("server berjalan", slog.String("port", port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.ShutdownWithContext(ctx)
}