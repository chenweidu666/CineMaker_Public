package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cinemaker/backend/api/routes"
	"github.com/cinemaker/backend/domain/models"
	"github.com/cinemaker/backend/infrastructure/database"
	"github.com/cinemaker/backend/infrastructure/storage"
	"github.com/cinemaker/backend/pkg/config"
	"github.com/cinemaker/backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logr := logger.NewLogger(cfg.App.Debug)
	defer logr.Sync()

	logr.Info("Starting Drama Generator API Server...")

	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logr.Fatal("Failed to connect to database", "error", err)
	}
	logr.Info("Database connected successfully")

	// 自动迁移数据库表结构
	if err := database.AutoMigrate(db); err != nil {
		logr.Fatal("Failed to migrate database", "error", err)
	}
	logr.Info("Database tables migrated successfully")

	// 兼容迁移：创建默认团队和默认管理员
	migrateDefaultTeam(db, logr, cfg)
	// 迁移旧默认邮箱 admin@localhost -> admin@example.com（修复邮箱格式校验）
	migrateAdminEmail(db, logr)

	// 初始化存储后端
	var store storage.Storage
	switch cfg.Storage.Type {
	case "cos":
		store, err = storage.NewCOSStorage(storage.COSConfig{
			Bucket:    cfg.Storage.COS.Bucket,
			Region:    cfg.Storage.COS.Region,
			SecretID:  cfg.Storage.COS.SecretID,
			SecretKey: cfg.Storage.COS.SecretKey,
			CDNURL:    cfg.Storage.COS.CDNURL,
		}, cfg.Storage.LocalPath) // local path used as temp dir for COS mode
		if err != nil {
			logr.Fatal("Failed to initialize COS storage", "error", err)
		}
		logr.Info("COS storage initialized", "bucket", cfg.Storage.COS.Bucket, "region", cfg.Storage.COS.Region)
	default:
		store, err = storage.NewLocalStorage(cfg.Storage.LocalPath, cfg.Storage.BaseURL)
		if err != nil {
			logr.Fatal("Failed to initialize local storage", "error", err)
		}
		logr.Info("Local storage initialized", "path", cfg.Storage.LocalPath)
	}

	if cfg.App.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := routes.SetupRouter(cfg, db, logr, store)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Minute,
		WriteTimeout: 10 * time.Minute,
	}

	go func() {
		logr.Infow("🚀 Server starting...",
			"port", cfg.Server.Port,
			"mode", gin.Mode())
		logr.Info("📍 Access URLs:")
		logr.Info(fmt.Sprintf("   Frontend:  http://localhost:%d", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   API:       http://localhost:%d/api/v1", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   Health:    http://localhost:%d/health", cfg.Server.Port))
		logr.Info("📁 Static files:")
		logr.Info(fmt.Sprintf("   Uploads:   http://localhost:%d/static", cfg.Server.Port))
		logr.Info(fmt.Sprintf("   Assets:    http://localhost:%d/assets", cfg.Server.Port))
		logr.Info("✅ Server is ready!")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logr.Fatal("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logr.Info("Shutting down server...")

	// 清理资源
	// CRITICAL FIX: Properly close database connection to prevent resource leaks
	// SQLite connections should be closed gracefully to avoid database lock issues
	sqlDB, err := db.DB()
	if err == nil {
		if err := sqlDB.Close(); err != nil {
			logr.Warnw("Failed to close database connection", "error", err)
		} else {
			logr.Info("Database connection closed")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logr.Fatal("Server forced to shutdown", "error", err)
	}

	logr.Info("Server exited")
}

func migrateDefaultTeam(db *gorm.DB, logr *logger.Logger, cfg *config.Config) {
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		return
	}

	logr.Info("No users found, creating default team and admin user...")

	team := &models.Team{Name: "默认团队"}
	if err := db.Create(team).Error; err != nil {
		logr.Warnw("Failed to create default team", "error", err)
		return
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminEmail == "" && cfg != nil && cfg.App.DefaultAdminEmail != "" {
		adminEmail = cfg.App.DefaultAdminEmail
	}
	if adminPassword == "" && cfg != nil && cfg.App.DefaultAdminPassword != "" {
		adminPassword = cfg.App.DefaultAdminPassword
	}
	if adminEmail != "" && adminPassword != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err == nil {
			admin := &models.User{
				Username:     "admin",
				Email:        adminEmail,
				PasswordHash: string(hash),
				Role:         "owner",
				TeamID:       &team.ID,
			}
			if err := db.Create(admin).Error; err == nil {
				team.OwnerID = admin.ID
				db.Save(team)
				logr.Infow("Default admin created", "email", adminEmail)
			}
		}
	}

	tables := []string{"dramas", "ai_service_configs", "assets", "character_libraries"}
	for _, table := range tables {
		result := db.Exec("UPDATE "+table+" SET team_id = ? WHERE team_id IS NULL", team.ID)
		if result.RowsAffected > 0 {
			logr.Infow("Backfilled team_id", "table", table, "rows", result.RowsAffected)
		}
	}
}

func migrateAdminEmail(db *gorm.DB, logr *logger.Logger) {
	result := db.Model(&models.User{}).Where("email = ?", "admin@localhost").Update("email", "admin@example.com")
	if result.RowsAffected > 0 {
		logr.Infow("Migrated admin email", "from", "admin@localhost", "to", "admin@example.com")
	}
}
