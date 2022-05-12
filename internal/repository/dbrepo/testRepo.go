package dbrepo

import (
	"errors"
	"fmt"

	"github.com/cepa995/go-web-template/internal/models"
)

// Authenticate authenticates the user.
func (m *testDBRepo) Authenticate(email string, testPassword string) (int64, string, error) {
	// Here we are mocking what is happening on DB level
	if email == "test@gmail.com" && testPassword == "password" {
		return 1, "", nil
	}
	return 0, "", errors.New("invalid email credentials")
}

// AllUsers retrieves list of all users from the database.
func (m *testDBRepo) AllUsers() ([]models.User, error) {
	var users []models.User
	return users, nil
}

//InsertUser inserts user into the database
func (m *testDBRepo) InsertUser(user models.User) (int64, error) {
	return 1, nil
}

// UpdateUser updates user in the database
func (m *testDBRepo) UpdateUser(user models.User) error {
	return nil
}

// GetUserByID retrieves user from the database by ID
func (m *testDBRepo) GetUserByID(userID int64) (models.User, error) {
	if userID == 1 {
		return models.User{
			ID:        1,
			FirstName: "Jon",
			LastName:  "Doe",
			Email:     "test@gmail.com",
			Password:  "password",
		}, nil
	}
	return models.User{}, fmt.Errorf("user with ID %d does not exist", userID)
}

// GetUserByEmail retrieves user from the database by email
func (m *testDBRepo) GetUserByEmail(email string) (models.User, error) {
	if email == "test@gmail.com" {
		return models.User{
			ID:        1,
			FirstName: "Jon",
			LastName:  "Doe",
			Email:     "test@gmail.com",
			Password:  "password",
		}, nil
	}
	return models.User{}, fmt.Errorf("user with email %s does not exist", email)
}

// UpdatePasswordForUser updates specified user's hashed password
func (m *testDBRepo) UpdatePasswordForUser(user models.User, hash string) error {
	return nil
}
