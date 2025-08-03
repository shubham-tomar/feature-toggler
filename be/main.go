package main

import (
	"github.com/shubham-tomar/feature-toggler/graphQl/resolver"
	"github.com/shubham-tomar/feature-toggler/graphQl/context"
	"github.com/shubham-tomar/feature-toggler/graphQl/generated"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// GraphQL handler
	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{
		Resolvers: &resolver.Resolver{},
	}))

	// Middleware: inject mock user into context
	r.POST("/query", func(c *gin.Context) {
		ctx := context.WithUser(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)
		srv.ServeHTTP(c.Writer, c.Request)
	})

	// GraphQL playground
	r.GET("/", func(c *gin.Context) {
		playground.Handler("GraphQL playground", "/query").ServeHTTP(c.Writer, c.Request)
	})

	r.Run(":8080")
}
