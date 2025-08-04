package resolver

import (
	"github.com/shubham-tomar/feature-toggler/db"
	"github.com/shubham-tomar/feature-toggler/graphQl/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct{
	projects []*model.Project
	Storage  db.Storage
}
