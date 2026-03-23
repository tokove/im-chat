package model

import (
	"backend/internal/db"
	"database/sql"
	"errors"
	"time"
)

type Private struct {
	ID        int64     `json:"id"`
	User1     int64     `json:"user1"`
	User2     int64     `json:"user2"`
	CreatedAt time.Time `json:"created_at"`
}

func GetPrivateByID(privateID int64) (*Private, error) {
	p := &Private{}
	row := db.DB.QueryRow(`
	 	SELECT id, user1_id, user2_id, created_at
		FROM privates
		WHERE id = ?
	`, privateID)

	err := row.Scan(&p.ID, &p.User1, &p.User2, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return p, nil
}

func GetPrivateByUsers(user1ID, user2ID int64) (*Private, error) {
	p := &Private{}

	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}
	row := db.DB.QueryRow(`
	 	SELECT id, user1_id, user2_id, created_at
		FROM privates
		WHERE user1_id = ? AND user2_id = ?
	`, user1ID, user2ID)

	err := row.Scan(&p.ID, &p.User1, &p.User2, &p.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return p, nil
}

func GetPrivatesForUser(userID int64) ([]Private, error) {
	privates := make([]Private, 0)

	rows, err := db.DB.Query(`
		SELECT id, user1_id, user2_id, created_at 
		FROM privates
		WHERE user1_id = ? OR user2_id = ?
		ORDER BY created_at DESC
	`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		p := Private{}
		if err := rows.Scan(&p.ID, &p.User1, &p.User2, &p.CreatedAt); err != nil {
			return nil, err
		}
		privates = append(privates, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return privates, nil
}

func CreatePrivate(user1ID, user2ID int64) (*Private, error) {
	if user1ID == user2ID {
		return nil, errors.New("Cannot create private chat with the same user")
	}

	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	existingPrivate, err := GetPrivateByUsers(user1ID, user2ID)
	if err != nil {
		return nil, err
	}
	if existingPrivate != nil {
		return nil, errors.New("Private already exist")
	}

	res, err := db.DB.Exec(`
		INSERT INTO privates (user1_id, user2_id) VALUES (?, ?)
	`, user1ID, user2ID)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	var createdAt time.Time
	if err := db.DB.QueryRow(`SELECT created_at FROM privates WHERE id = ?`, id).Scan(&createdAt); err != nil {
		return nil, err
	}

	return &Private{
		ID:        id,
		User1:     user1ID,
		User2:     user2ID,
		CreatedAt: createdAt,
	}, nil
}