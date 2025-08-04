package db

import (
	"context"
	"github.com/shubham-tomar/feature-toggler/graphQl/model"
)

// Storage defines the interface for database operations
type Storage interface {
	// Connection management
	Connect() error
	Close() error
	Ping(ctx context.Context) error
	
	// User operations
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByID(ctx context.Context, id string) (*model.User, error)
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	// DeleteUser(ctx context.Context, id string) error
	
	// Project operations
	CreateProject(ctx context.Context, user *model.User, name string) (*model.Project, error)
	GetProjectByID(ctx context.Context, id string) (*model.Project, error)
	GetUserProjects(ctx context.Context, user *model.User) ([]*model.Project, error)
	GetProjects(ctx context.Context) ([]*model.Project, error)
	UpdateProject(ctx context.Context, project *model.Project) error
	DeleteProject(ctx context.Context, id string) error
	
	// Project membership operations
	AddProjectMember(ctx context.Context, membership *model.ProjectUser) error
	UpdateProjectMemberRole(ctx context.Context, membershipID string, role model.Role) error
	RemoveProjectMember(ctx context.Context, membershipID string) error
	GetProjectMembers(ctx context.Context, projectID string) ([]*model.ProjectUser, error)
	
	// Feature flag operations
	CreateFeatureFlag(ctx context.Context, flag *model.FeatureFlag, initialStates []*model.ToggleState) error
	GetFeatureFlagByID(ctx context.Context, id string) (*model.FeatureFlag, error)
	GetFeatureFlagByKey(ctx context.Context, key string) (*model.FeatureFlag, error)
	GetProjectFeatureFlags(ctx context.Context, projectID string) ([]*model.FeatureFlag, error)
	UpdateFeatureFlag(ctx context.Context, flag *model.FeatureFlag) error
	DeleteFeatureFlag(ctx context.Context, id string) error
	
	// Toggle state operations
	GetFeatureFlagStates(ctx context.Context, flagID string) ([]*model.ToggleState, error)
	UpdateFeatureFlagState(ctx context.Context, state *model.ToggleState) error
}

// StorageFactory creates new storage instances
type StorageFactory interface {
	NewStorage() (Storage, error)
}
