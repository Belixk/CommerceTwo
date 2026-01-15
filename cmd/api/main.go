package main

import (
	"context"
	"log"
	"time"

	"github.com/Belixk/CommerceTwo/config"
	"github.com/Belixk/CommerceTwo/internal/handlers"
	"github.com/Belixk/CommerceTwo/internal/repositories"
	"github.com/Belixk/CommerceTwo/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.Load()
	var db *sqlx.DB
	var err error

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

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	userRepo := repositories.NewUserRepository(db)
	userCache := repositories.NewUserCache(rdb)
	userService := services.NewUserService(userRepo, userCache)
	userHandler := handlers.NewUserHandler(userService)

	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUser)
		}
	}

	log.Printf("Server starting on :%s", cfg.AppPort)
	r.Run(":" + cfg.AppPort)
}
