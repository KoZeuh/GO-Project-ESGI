package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
)

// UserRepository encapsule l'accès à la table "users".
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create insère un nouvel utilisateur et renvoie son identifiant généré.
func (r *UserRepository) Create(u *models.User) (int64, error) {
	res, err := r.db.Exec(
		`INSERT INTO users (username, password_hash) VALUES (?, ?)`,
		u.Username, u.PasswordHash,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			return 0, ErrDuplicate
		}
		return 0, fmt.Errorf("création utilisateur : %w", err)
	}
	return res.LastInsertId()
}

// FindByUsername retourne l'utilisateur portant ce nom, ou ErrNotFound.
func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE username = ?`,
		username,
	)
	return scanUser(row)
}

// FindByID retourne l'utilisateur portant cet identifiant, ou ErrNotFound.
func (r *UserRepository) FindByID(id int64) (*models.User, error) {
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, created_at FROM users WHERE id = ?`,
		id,
	)
	return scanUser(row)
}

func scanUser(row *sql.Row) (*models.User, error) {
	var u models.User
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("lecture utilisateur : %w", err)
	}
	return &u, nil
}
