package model

import (
	"backend/internal/db"
	"backend/internal/middleware"
	"database/sql"
	"errors"
	"time"
)

type User struct {
	ID                   int64      `json:"id"`
	Name                 string     `json:"name"`
	Email                string     `json:"email"`
	Password             string     `json:"-"`
	RefreshTokenWeb      *string    `json:"-"`
	RefreshTokenWebAt    *time.Time `json:"-"`
	RefreshTokenMobile   *string    `json:"-"`
	RefreshTokenMobileAt *time.Time `json:"-"`
	CreatedAt            time.Time  `json:"created_at"`
}

func GetUserByEmail(email string) (*User, error) {
	u := &User{}

	row := db.DB.QueryRow(
		`SELECT id, name, email, password, refresh_token_web, refresh_token_web_at, refresh_token_mobile, refresh_token_mobile_at, created_at
		FROM users
		WHERE email = ?`,
		email,
	)

	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.RefreshTokenWeb,
		&u.RefreshTokenWebAt,
		&u.RefreshTokenMobile,
		&u.RefreshTokenMobileAt,
		&u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

func CreateUserByEmail(name, email, password string) (*User, error) {
	res, err := db.DB.Exec(`
		INSERT INTO users (name, email, password) 
		VALUES (?, ?, ?)`, name, email, password)
	if err != nil {
		return nil, err
	}

	id, _ := res.LastInsertId()
	createdAt := time.Now()

	return &User{ID: id, Name: name, Email: email, CreatedAt: createdAt}, nil
}

func UpdateUserRefreshToken(userID int64, platform, refreshToken string) error {
	switch platform {
	case middleware.PlatformWeb:
		_, err := db.DB.Exec(`
			UPDATE users
			SET refresh_token_web = ?, refresh_token_web_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, refreshToken, userID)
		return err

	case middleware.PlatformMobile:
		_, err := db.DB.Exec(`
			UPDATE users
			SET refresh_token_mobile = ?, refresh_token_mobile_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`, refreshToken, userID)
		return err

	default:
		return errors.New("Invalid platform")
	}
}

func DeleteUserRefreshToken(userID int64, platform string) error {
	switch platform {
	case middleware.PlatformWeb:
		_, err := db.DB.Exec(`
			UPDATE users
			SET refresh_token_web = NULL, refresh_token_web_at = NULL
			WHERE id = ?
		`, userID)
		return err

	case middleware.PlatformMobile:
		_, err := db.DB.Exec(`
			UPDATE users
			SET refresh_token_mobile = NULL, refresh_token_mobile_at = NULL
			WHERE id = ?
		`, userID)
		return err

	default:
		return errors.New("Invalid platform")
	}
}

func GetUserByRefreshToken(refreshToken, platform string) (*User, error) {
	u := &User{}

	var row *sql.Row

	switch platform {
	case middleware.PlatformWeb:
		row = db.DB.QueryRow(
			`SELECT id, name, email, password, refresh_token_web, refresh_token_web_at, refresh_token_mobile, refresh_token_mobile_at, created_at
			FROM users
			WHERE refresh_token_web = ?`,
			refreshToken,
		)
	case middleware.PlatformMobile:
		row = db.DB.QueryRow(
			`SELECT id, name, email, password, refresh_token_web, refresh_token_web_at, refresh_token_mobile, refresh_token_mobile_at, created_at
			FROM users
			WHERE refresh_token_mobile = ?`,
			refreshToken,
		)
	default:
		return nil, errors.New("invalid platform")
	}

	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.RefreshTokenWeb,
		&u.RefreshTokenWebAt,
		&u.RefreshTokenMobile,
		&u.RefreshTokenMobileAt,
		&u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}

func GetUserByID(id int64) (*User, error) {
	u := &User{}

	row := db.DB.QueryRow(
		`SELECT id, name, email, password, refresh_token_web, refresh_token_web_at, refresh_token_mobile, refresh_token_mobile_at, created_at
		FROM users
		WHERE id = ?`,
		id,
	)

	err := row.Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.Password,
		&u.RefreshTokenWeb,
		&u.RefreshTokenWebAt,
		&u.RefreshTokenMobile,
		&u.RefreshTokenMobileAt,
		&u.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return u, nil
}
