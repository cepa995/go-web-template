package dbrepo

import (
	"context"
	"errors"
	"time"

	"github.com/cepa995/go-web-template/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UpdatePasswordForUser updates specified user's hashed password
func (m *postgresDBRepo) UpdatePasswordForUser(user models.User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	stmt := `update users set password = $1 where id = $2`
	_, err := m.DB.ExecContext(ctx, stmt, hash, user.ID)
	if err != nil {
		return err
	}
	return nil
}

// Authenticate authenticates the user.
func (m *postgresDBRepo) Authenticate(email string, testPassword string) (int64, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	var userID int64
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = $1", email)
	err := row.Scan(&userID, &hashedPassword)
	if err != nil {
		return userID, "", err
	}

	// Built-in package fro comparing hashed password pulled from DB and password that user typed into the form.
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(testPassword))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, "", errors.New("incorrect password")
	} else if err != nil {
		return 0, "", err
	}

	return userID, hashedPassword, nil
}

// AllUsers retrieves list of all users from the database.
func (m *postgresDBRepo) AllUsers() ([]models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	var users []models.User
	query := `
		select
			*
		from users
	`
	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Email,
			&user.Password,
			&user.AccessLevel,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return users, err
		}
		users = append(users, user)
	}

	return users, nil
}

//InsertUser inserts user into the database
func (m *postgresDBRepo) InsertUser(user models.User) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	var newID int64
	query := `insert into users (first_name, last_name, email, password, access_level, blocked, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8) returning id`
	err := m.DB.QueryRowContext(ctx, query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.Password,
		user.AccessLevel,
		false,
		time.Now(),
		time.Now(),
	).Scan(&newID)
	if err != nil {
		return 0, err
	}

	return newID, nil
}

// UpdateUser updates user in the database
func (m *postgresDBRepo) UpdateUser(user models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	query := `
		update users set first_name = $1, last_name = $2, email = $3, access_level = $4, updated_at = $5
	`
	_, err := m.DB.ExecContext(ctx, query,
		user.FirstName,
		user.LastName,
		user.Email,
		user.AccessLevel,
		time.Now(),
	)
	if err != nil {
		return err
	}

	return nil
}

// GetUser retrieves user from the database by ID
func (m *postgresDBRepo) GetUserByID(userID int64) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	var user models.User

	query := `
			select 
				*
			from 
				users u
			where u.id = $1;
	`

	err := m.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.AccessLevel,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return user, err
	}

	return user, nil
}

// GetUserByEmail retrieves user from the database by email
func (m *postgresDBRepo) GetUserByEmail(email string) (models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	var user models.User
	query := `
			select 
				*
			from 
				users u
			where u.email= $1;
	`

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.AccessLevel,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return user, err
	}

	return user, nil
}
