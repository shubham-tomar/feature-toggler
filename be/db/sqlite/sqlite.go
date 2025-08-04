package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shubham-tomar/feature-toggler/db"
	"github.com/shubham-tomar/feature-toggler/graphQl/model"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	db     *sql.DB
	dbPath string
}

// GetDB returns the underlying database connection
func (s *SQLiteStorage) GetDB() *sql.DB {
	return s.db
}

type SQLiteFactory struct {
	DBPath string
}

func (f *SQLiteFactory) NewStorage() (db.Storage, error) {
	return &SQLiteStorage{
		dbPath: f.DBPath,
	}, nil
}

func (s *SQLiteStorage) Connect() error {
	database, err := sql.Open("sqlite3", s.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	s.db = database
	return nil
}

func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

func (s *SQLiteStorage) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// User operations
func (s *SQLiteStorage) CreateUser(ctx context.Context, user *model.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?)`,
		user.ID, user.Name, user.Email, user.CreatedAt, user.UpdatedAt,
	)

	return err
}

func (s *SQLiteStorage) GetUserByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, email, created_at, updated_at FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SQLiteStorage) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, email, created_at, updated_at FROM users WHERE email = ?`,
		email,
	).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SQLiteStorage) UpdateUser(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now()

	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET name = ?, email = ?, updated_at = ? WHERE id = ?`,
		user.Name, user.Email, user.UpdatedAt, user.ID,
	)

	return err
}

// Project operations

func (s *SQLiteStorage) GetProjects(ctx context.Context) ([]*model.Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, created_at, updated_at FROM projects`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}

		// Get project members
		p.Members, err = s.GetProjectMembers(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting project members: %w", err)
		}

		projects = append(projects, &p)
	}

	return projects, nil
}

func (s *SQLiteStorage) CreateProject(ctx context.Context, user *model.User, name string) (*model.Project, error) {
	projectID := uuid.New().String()
	projectUserID := uuid.New().String()

	// Create timestamps for consistency
	now := time.Now()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`INSERT INTO projects (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`, projectID, name, now, now)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(`INSERT INTO project_users (id, user_id, project_id, role, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		projectUserID, user.ID, projectID, "ADMIN", now, now)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &model.Project{
		ID:        projectID,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Members: []*model.ProjectUser{
			{
				ID:      projectUserID,
				User:    user,
				Project: nil, // Will be set by GraphQL resolver
				Role:    model.RoleAdmin,
			},
		},
	}, nil
}

func (s *SQLiteStorage) GetProjectByID(ctx context.Context, id string) (*model.Project, error) {
	var project model.Project
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at, updated_at FROM projects WHERE id = ?`,
		id,
	).Scan(&project.ID, &project.Name, &project.CreatedAt, &project.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("project not found")
	}

	if err != nil {
		return nil, err
	}

	// Get project members
	project.Members, err = s.GetProjectMembers(ctx, project.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting project members: %w", err)
	}

	return &project, nil
}

func (s *SQLiteStorage) GetUserProjects(ctx context.Context, user *model.User) ([]*model.Project, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT p.id, p.name, p.created_at, p.updated_at 
		FROM projects p
		JOIN project_users pu ON p.id = pu.project_id
		WHERE pu.user_id = ?`,
		user.ID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*model.Project
	for rows.Next() {
		var p model.Project
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}

		// Get project members
		p.Members, err = s.GetProjectMembers(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting project members: %w", err)
		}

		projects = append(projects, &p)
	}

	return projects, nil
}

func (s *SQLiteStorage) UpdateProject(ctx context.Context, project *model.Project) error {
	project.UpdatedAt = time.Now()

	_, err := s.db.ExecContext(ctx,
		`UPDATE projects SET name = ?, updated_at = ? WHERE id = ?`,
		project.Name, project.UpdatedAt, project.ID,
	)

	return err
}

func (s *SQLiteStorage) DeleteProject(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM projects WHERE id = ?`, id)
	return err
}

// Project membership operations
func (s *SQLiteStorage) AddProjectMember(ctx context.Context, membership *model.ProjectUser) error {
	if membership.ID == "" {
		membership.ID = uuid.New().String()
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO project_users (id, user_id, project_id, role, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		membership.ID, membership.User.ID, membership.Project.ID, membership.Role,
		time.Now(), time.Now(),
	)

	return err
}

func (s *SQLiteStorage) UpdateProjectMemberRole(ctx context.Context, membershipID string, role model.Role) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE project_users SET role = ?, updated_at = ? WHERE id = ?`,
		role, time.Now(), membershipID,
	)

	return err
}

func (s *SQLiteStorage) RemoveProjectMember(ctx context.Context, membershipID string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM project_users WHERE id = ?`,
		membershipID,
	)

	return err
}

func (s *SQLiteStorage) GetProjectMembers(ctx context.Context, projectID string) ([]*model.ProjectUser, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT pu.id, pu.user_id, pu.project_id, pu.role 
		FROM project_users pu WHERE pu.project_id = ?`,
		projectID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*model.ProjectUser
	for rows.Next() {
		var m model.ProjectUser
		var userID, projectID string

		if err := rows.Scan(&m.ID, &userID, &projectID, &m.Role); err != nil {
			return nil, err
		}

		m.Project = &model.Project{ID: projectID}

		user, err := s.GetUserByID(ctx, userID)
		if err != nil {
			m.User = &model.User{ID: userID}
		} else {
			m.User = user
		}

		members = append(members, &m)
	}

	return members, nil
}

// Feature flag operations
func (s *SQLiteStorage) CreateFeatureFlag(ctx context.Context, flag *model.FeatureFlag, initialStates []*model.ToggleState) error {
	if flag.ID == "" {
		flag.ID = uuid.New().String()
	}

	now := time.Now()
	flag.CreatedAt = now
	flag.UpdatedAt = now

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var projectID, createdByID string
	var description string

	if flag.Project != nil {
		projectID = flag.Project.ID
	}

	if flag.CreatedBy != nil {
		createdByID = flag.CreatedBy.ID
	}

	if flag.Description != nil {
		description = *flag.Description
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO feature_flags (id, key, name, description, project_id, created_by_id, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		flag.ID, flag.Key, flag.Name, description, projectID, createdByID, flag.CreatedAt, flag.UpdatedAt,
	)

	if err != nil {
		tx.Rollback()
		return err
	}

	for _, state := range initialStates {
		state.ID = uuid.New().String()

		if state.FeatureFlag == nil {
			state.FeatureFlag = &model.FeatureFlag{ID: flag.ID}
		} else {
			state.FeatureFlag.ID = flag.ID
		}

		state.UpdatedAt = now

		var updatedByID string
		if state.UpdatedBy != nil {
			updatedByID = state.UpdatedBy.ID
		}

		_, err = tx.ExecContext(ctx,
			`INSERT INTO toggle_states (id, feature_flag_id, environment, enabled, updated_by_id, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?)`,
			state.ID, state.FeatureFlag.ID, state.Environment, state.Enabled, updatedByID, state.UpdatedAt,
		)

		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *SQLiteStorage) GetFeatureFlagByID(ctx context.Context, id string) (*model.FeatureFlag, error) {
	var flag model.FeatureFlag
	var projectID, createdByID string
	var description sql.NullString

	err := s.db.QueryRowContext(ctx,
		`SELECT id, key, name, description, project_id, created_by_id, created_at, updated_at 
		FROM feature_flags WHERE id = ?`,
		id,
	).Scan(&flag.ID, &flag.Key, &flag.Name, &description, &projectID, &createdByID, &flag.CreatedAt, &flag.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("feature flag not found")
	}

	if err != nil {
		return nil, err
	}

	// Handle description
	if description.Valid {
		desc := description.String
		flag.Description = &desc
	}

	// Get project data
	if projectID != "" {
		project, err := s.GetProjectByID(ctx, projectID)
		if err != nil {
			return nil, fmt.Errorf("error getting project: %w", err)
		}
		flag.Project = project
	}

	// Get created by user data
	if createdByID != "" {
		user, err := s.GetUserByID(ctx, createdByID)
		if err != nil {
			return nil, fmt.Errorf("error getting user: %w", err)
		}
		flag.CreatedBy = user
	}

	// Get toggle states
	flag.States, err = s.GetFeatureFlagStates(ctx, flag.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting toggle states: %w", err)
	}

	return &flag, nil
}

func (s *SQLiteStorage) GetFeatureFlagByKey(ctx context.Context, key string) (*model.FeatureFlag, error) {
	var flag model.FeatureFlag
	err := s.db.QueryRowContext(ctx,
		`SELECT id, key, name, description, project_id, created_by_id, created_at, updated_at 
		FROM feature_flags WHERE key = ?`,
		key,
	).Scan(&flag.ID, &flag.Key, &flag.Name, &flag.Description, &flag.Project.ID,
		&flag.CreatedBy.ID, &flag.CreatedAt, &flag.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, errors.New("feature flag not found")
	}

	if err != nil {
		return nil, err
	}

	return &flag, nil
}

func (s *SQLiteStorage) GetProjectFeatureFlags(ctx context.Context, projectID string) ([]*model.FeatureFlag, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, key, name, description, project_id, created_by_id, created_at, updated_at 
		FROM feature_flags WHERE project_id = ?`,
		projectID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []*model.FeatureFlag
	for rows.Next() {
		var f model.FeatureFlag
		if err := rows.Scan(&f.ID, &f.Key, &f.Name, &f.Description, &f.Project.ID,
			&f.CreatedBy.ID, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		flags = append(flags, &f)
	}

	return flags, nil
}

func (s *SQLiteStorage) UpdateFeatureFlag(ctx context.Context, flag *model.FeatureFlag) error {
	flag.UpdatedAt = time.Now()

	_, err := s.db.ExecContext(ctx,
		`UPDATE feature_flags SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		flag.Name, flag.Description, flag.UpdatedAt, flag.ID,
	)

	return err
}

func (s *SQLiteStorage) DeleteFeatureFlag(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM feature_flags WHERE id = ?`, id)
	return err
}

// Toggle state operations
func (s *SQLiteStorage) GetFeatureFlagStates(ctx context.Context, flagID string) ([]*model.ToggleState, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, feature_flag_id, environment, enabled, updated_by_id, updated_at 
		FROM toggle_states WHERE feature_flag_id = ?`,
		flagID,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []*model.ToggleState
	for rows.Next() {
		var ts model.ToggleState
		var featureFlagID, updatedByID string

		if err := rows.Scan(&ts.ID, &featureFlagID, &ts.Environment, &ts.Enabled, &updatedByID, &ts.UpdatedAt); err != nil {
			return nil, err
		}

		// Initialize nested objects
		ts.FeatureFlag = &model.FeatureFlag{ID: featureFlagID}

		// Get user data if available
		if updatedByID != "" {
			user, err := s.GetUserByID(ctx, updatedByID)
			if err != nil {
				return nil, fmt.Errorf("error getting user: %w", err)
			}
			ts.UpdatedBy = user
		}

		states = append(states, &ts)
	}

	return states, nil
}

func (s *SQLiteStorage) UpdateFeatureFlagState(ctx context.Context, state *model.ToggleState) error {
	state.UpdatedAt = time.Now()

	var updatedByID string
	if state.UpdatedBy != nil {
		updatedByID = state.UpdatedBy.ID
	}

	var featureFlagID string
	if state.FeatureFlag != nil {
		featureFlagID = state.FeatureFlag.ID
	}

	_, err := s.db.ExecContext(ctx,
		`UPDATE toggle_states 
		SET enabled = ?, updated_by_id = ?, updated_at = ? 
		WHERE id = ? AND feature_flag_id = ? AND environment = ?`,
		state.Enabled, updatedByID, state.UpdatedAt,
		state.ID, featureFlagID, state.Environment,
	)

	return err
}
