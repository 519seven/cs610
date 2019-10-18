package sqlite3

import (
	"github.com/519seven/cs610/battleship/pkg/models"
	"database/sql"
	"fmt"
	"golang.org/x/xerrors"
	"strings"
	"time"
)

type PlayerModel struct {
	DB *sql.DB
}

// authenticate player
func (m *PlayerModel) Authenticate(email, password string) (int, error) {
	return 0, nil
}

// get player information
func (m *PlayerModel) Get(rowid int) (*models.Player, error) {
	stmt := `SELECT rowid, screenName, emailAddress, lastLogin FROM Players WHERE rowid = ?`
	s := &models.Player{}
	err := m.DB.QueryRow(stmt, rowid).Scan(&s.ID, &s.ScreenName, &s.EmailAddress, &s.LastLogin)
	if err != nil {
		fmt.Println("[ERROR] Error encountered:", err.Error())
		return nil, err
	}
	return s, nil
}

// insert new player
func (m *PlayerModel) Insert(screenName string, emailAddress string, password string) (int, error) {
	hashedPassword := "!Q@W#E$R%T^Y&U*I(O)P"
	stmt := `INSERT INTO Players (screenName, emailAddress, hashedPassword, isActive, lastLogin) VALUES (?, ?, ?, ?, ?)`
	result, err := m.DB.Exec(stmt, screenName, emailAddress, hashedPassword, 1, time.Now())
	if err != nil {
		fmt.Println(err.Error())
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			return 0, models.ErrDuplicateEmail
		} else {
			return 0, err
		}
	}
	rowid, err := result.LastInsertId() // confirmed! sqlite3 driver has this functionality!
	if err != nil {
		return 0, err
	}
	// id has type int64; convert to int type before returning
	return int(rowid), nil
}

// list players
func (m *PlayerModel) List() ([]*models.Player, error) {
	stmt := `SELECT rowid, screenName, isActive, lastLogin FROM Players`
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

// update player
func (m *PlayerModel) Update(id int, emailAddress string) (int, error) {
	stmt := `UPDATE Players SET emailAddress = ?, lastLogin = ? WHERE rowid = ?`
	_, err := m.DB.Exec(stmt, emailAddress, time.Now(), id)
	if err != nil {
		if xerrors.Is(err, sql.ErrNoRows) {
			return id, models.ErrNoRecord
		}
	}
	return id, err
}
