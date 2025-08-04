package main

import (
	"context"
	"log"
	"time"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"github.com/shubham-tomar/feature-toggler/db/sqlite"
	userctx "github.com/shubham-tomar/feature-toggler/graphQl/context"
	"github.com/shubham-tomar/feature-toggler/graphQl/generated"
	"github.com/shubham-tomar/feature-toggler/graphQl/resolver"
	"github.com/shubham-tomar/feature-toggler/utils"
	"github.com/shubham-tomar/feature-toggler/graphQl/model"
)

func main() {
	r := gin.Default()
	
	dbPath := utils.GetEnv("DB_PATH", "./feature-toggler.db")
	port := utils.GetEnv("PORT", "8080")

	// Initialize SQLite storage
	storageFactory := &sqlite.SQLiteFactory{DBPath: dbPath}
	sqliteStorage, err := storageFactory.NewStorage()
	if err != nil {
		log.Fatalf("Failed to create SQLite storage: %v", err)
	}
	
	if err := sqliteStorage.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer sqliteStorage.Close()
	
	// Run migrations using the SQLite-specific DB handle
	if err := sqlite.Migrate(sqliteStorage.(*sqlite.SQLiteStorage).GetDB()); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	
	// Insert mock user if it doesn't exist
	ctx := context.Background()
	
	// Create a mock user object with proper timestamps
	mockUser := &model.User{
		ID:        userctx.MockUser.ID,
		Name:      userctx.MockUser.Name,
		Email:     userctx.MockUser.Email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	// Check if user exists
	_, err = sqliteStorage.GetUserByID(ctx, mockUser.ID)
	if err != nil {
		// User doesn't exist, create it
		if err := sqliteStorage.CreateUser(ctx, mockUser); err != nil {
			log.Printf("Warning: Failed to create mock user: %v", err)
		} else {
			log.Printf("Mock user created with ID: %s, Name: %s", mockUser.ID, mockUser.Name)
		}
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolver.Resolver{
			Storage: sqliteStorage,
		},
	}))

	r.POST("/query", func(c *gin.Context) {
		ctx := userctx.WithUser(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		srv.ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/", func(c *gin.Context) {
		playground.Handler("GraphQL playground", "/query").ServeHTTP(c.Writer, c.Request)
	})

	log.Printf("Server running on :%s", port)
	r.Run(":" + port)
}