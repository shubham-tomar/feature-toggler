package context

import (
	"context"
	"github.com/shubham-tomar/feature-toggler/graphQl/model"
)

type ctxKey string

const userKey = ctxKey("user")

// Mock User
var MockUser = &model.User{
	ID:    "1",
	Name:  "tomar",
	Email: "tomar@pixis.ai",
}

func WithUser(ctx context.Context) context.Context {
	return context.WithValue(ctx, userKey, MockUser)
}

func GetUser(ctx context.Context) *model.User {
	user, _ := ctx.Value(userKey).(*model.User)
	return user
}


