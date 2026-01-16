package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Belixk/CommerceTwo/config"
	"github.com/Belixk/CommerceTwo/internal/handlers"
	"github.com/Belixk/CommerceTwo/internal/pkg/hash"
	"github.com/Belixk/CommerceTwo/internal/repositories"
	"github.com/Belixk/CommerceTwo/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	var db *sqlx.DB
	var err error
	// 5 попыток подключения к бд
	for i := 0; i < 5; i++ {
		log.Printf("Attemptin to connect to PostgreSQL #%d\n", i+1)
		db, err = sqlx.Connect("postgres", cfg.GetDBDSN())
		if err == nil {
			break
		}
		log.Printf("Postgres not ready yet: %v\n", err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("Postgres connection failed after 5 attempts: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL")
	defer db.Close()

	runMigrations(db)

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	userCache := repositories.NewUserCache(rdb)
	hasher := &hash.BcryptHasher{}
	userService := services.NewUserService(userRepo, userCache, hasher)
	userHandler := handlers.NewUserHandler(userService)

	orderRepo := repositories.NewOrderRepository(db)
	orderCache := repositories.NewOrderCache(rdb)

	orderService := services.NewOrderService(orderRepo, orderCache)
	orderHandler := handlers.NewOrderHandler(orderService)

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}
		orders := v1.Group("/orders")
		{
			orders.POST("/", orderHandler.CreateOrder)
			orders.GET("/:id", orderHandler.GetOrderByID)
			orders.GET("/user/:user_id", orderHandler.GetOrdersByUserID)
			orders.PUT("/:id", orderHandler.UpdateOrder)
			orders.DELETE("/:id", orderHandler.DeleteOrder)
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.AppPort,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Printf("Server starting on :%s", cfg.AppPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to Shutdown:", err)
	}
	log.Println("Closing database connections...")
	db.Close()
	rdb.Close()

	log.Println("Server exiting")
}

func runMigrations(db *sqlx.DB) {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		log.Fatalf("could not create migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", driver)
	if err != nil {
		log.Fatalf("could not create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Database is up to date (no migrations to apply)")
		} else {
			log.Fatalf("could not run up migrations: %v", err)
		}
	} else {
		log.Println("Migrations applied successfully!")
	}
}
