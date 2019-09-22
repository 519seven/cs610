package sqlite3

import (
	"519seven/battleship/pkg/models"
	"database/sql"
	"errors"
	"time"
)

type PlayerModel struct {
	DB *sql.DB
}

func (m *PlayerModel) Get(id int) (*models.Player, error) {
	stmt := `SELECT id, screenName, lastLogin FROM Users WHERE id = ?`
	s := &models.Player{}
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.ScreenName, &s.LastLogin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}
func (m *PlayerModel) Insert(screenName string) (int, error) {
	stmt := `INSERT INTO Users (screenName, lastLogin) VALUES (?, ?)`
	result, err := m.DB.Exec(stmt, screenName, time.Now())
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId() // confirmed! sqlite3 driver has this functionality!
	if err != nil {
		return 0, err
	}
	// id has type int64; convert to int type before returning
	return int(id), nil
}
func (m *PlayerModel) List() ([]*models.Player, error) {
	stmt := `SELECT id, screenName, isActive, lastLogin FROM Users`
	rows, err := m.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := []*models.Player{}

	for rows.Next() {
		s := &models.Player{}
		err = rows.Scan(&s.ID, &s.ScreenName, &s.IsActive, &s.LastLogin)
		if err != nil {
			return nil, err
		}
		players = append(players, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return players, nil
}
func (m *PlayerModel) Update(id int, screenName string) (int, error) {
	stmt := `UPDATE Users SET screenName = ?, lastLogin = ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, screenName, time.Now(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return id, models.ErrNoRecord
		}
	}
	return id, err
}
