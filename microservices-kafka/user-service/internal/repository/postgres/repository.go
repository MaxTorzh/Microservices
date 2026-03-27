package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"user-service/internal/domain"

	_ "github.com/lib/pq"
)

type Repository struct {
    db *sql.DB
}

func NewRepository(connStr string) (*Repository, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }
    
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return &Repository{db: db}, nil
}

func (r *Repository) Ping() error {
    return r.db.Ping()
}

func (r *Repository) Close() error {
    return r.db.Close()
}

func (r *Repository) Create(user domain.User) error {
    query := `
        INSERT INTO users (id, email, name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
    `
    
    _, err := r.db.Exec(query, user.ID, user.Email, user.Name, user.CreatedAt, user.UpdatedAt)
    if err != nil {
        if isUniqueViolation(err) {
            return domain.ErrEmailExists
        }
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    return nil
}

func (r *Repository) GetByID(id string) (domain.User, error) {
    query := `
        SELECT id, email, name, created_at, updated_at
        FROM users
        WHERE id = $1
    `
    
    var user domain.User
    err := r.db.QueryRow(query, id).Scan(
        &user.ID,
        &user.Email,
        &user.Name,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return domain.User{}, domain.ErrUserNotFound
        }
        return domain.User{}, fmt.Errorf("failed to get user: %w", err)
    }
    
    return user, nil
}

func (r *Repository) GetByEmail(email string) (domain.User, error) {
    query := `
        SELECT id, email, name, created_at, updated_at
        FROM users
        WHERE email = $1
    `
    
    var user domain.User
    err := r.db.QueryRow(query, email).Scan(
        &user.ID,
        &user.Email,
        &user.Name,
        &user.CreatedAt,
        &user.UpdatedAt,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return domain.User{}, domain.ErrUserNotFound
        }
        return domain.User{}, fmt.Errorf("failed to get user by email: %w", err)
    }
    
    return user, nil
}

func (r *Repository) GetAll() ([]domain.User, error) {
    query := `
        SELECT id, email, name, created_at, updated_at
        FROM users
        ORDER BY created_at DESC
    `
    
    rows, err := r.db.Query(query)
    if err != nil {
        return nil, fmt.Errorf("failed to query users: %w", err)
    }
    defer rows.Close()
    
    var users []domain.User
    for rows.Next() {
        var user domain.User
        if err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt, &user.UpdatedAt); err != nil {
            return nil, fmt.Errorf("failed to scan user: %w", err)
        }
        users = append(users, user)
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error iterating rows: %w", err)
    }
    
    return users, nil
}

func (r *Repository) Update(id string, user domain.User) error {
    query := `
        UPDATE users
        SET email = $1, name = $2, updated_at = $3
        WHERE id = $4
    `
    
    result, err := r.db.Exec(query, user.Email, user.Name, user.UpdatedAt, id)
    if err != nil {
        if isUniqueViolation(err) {
            return domain.ErrEmailExists
        }
        return fmt.Errorf("failed to update user: %w", err)
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rows == 0 {
        return domain.ErrUserNotFound
    }
    
    return nil
}

func (r *Repository) Delete(id string) error {
    query := `DELETE FROM users WHERE id = $1`
    
    result, err := r.db.Exec(query, id)
    if err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }
    
    rows, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get rows affected: %w", err)
    }
    
    if rows == 0 {
        return domain.ErrUserNotFound
    }
    
    return nil
}

func isUniqueViolation(err error) bool {
    if err == nil {
        return false
    }
    return strings.Contains(err.Error(), "duplicate key value violates unique constraint")
}