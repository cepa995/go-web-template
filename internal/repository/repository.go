package repository

import "github.com/cepa995/go-web-template/internal/models"

// DatabaseRepo interface which specifies set of operations for communicating with the database.
type DatabaseRepo interface {
	// User model functions
	AllUsers() ([]models.User, error)
	GetUserByID(userID int64) (models.User, error)
	GetUserByEmail(email string) (models.User, error)
	InsertUser(user models.User) (int64, error)
	UpdateUser(user models.User) error
	UpdatePasswordForUser(user models.User, newHash string) error
	Authenticate(email string, testPassword string) (int64, string, error)
}
