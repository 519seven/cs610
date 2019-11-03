package sqlite3

import (
	"github.com/519seven/cs610/battleship/pkg/models"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/xerrors"
	"strings"
	"time"
	_ "github.com/mattn/go-sqlite3"
)

type PlayerModel struct {
	DB *sql.DB
}

// authenticate player
func (m *PlayerModel) Authenticate(screenName, password string) (int, error) {
	var rowid int
	var hashedPassword []byte
	stmt := "SELECT rowid, hashedPassword FROM players WHERE screenName = ?"
	row := m.DB.QueryRow(stmt, screenName)
	err := row.Scan(&rowid, &hashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		} 
	}
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return 0, models.ErrInvalidCredentials
		} else {
			return 0, err
		}
	}
	// Otherwise, the password is correct. Return the user ID.
	return rowid, nil
}

// get player information
func (m *PlayerModel) Get(rowid int) (*models.Player, error) {
	p := &models.Player{}

	stmt := `SELECT rowid, screenName, emailAddress, lastLogin FROM Players WHERE rowid = ?`
	err := m.DB.QueryRow(stmt, rowid).Scan(&p.ID, &p.ScreenName, &p.EmailAddress, &p.LastLogin)
	if err != nil {
		fmt.Println("ERROR - ", err.Error())
		return nil, err
	}
	fmt.Println("INFO - Returning player information to authenticate middleware")
	return p, nil
}

// insert new player
func (m *PlayerModel) Insert(screenName string, emailAddress string, password string) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 13)
	if err != nil {
		fmt.Println("ERROR - Error generating hashed password")
		return 0, err
	}

	stmt := `INSERT INTO Players (screenName, emailAddress, hashedPassword, created, loggedIn, lastLogin) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := m.DB.Exec(stmt, screenName, emailAddress, hashedPassword, time.Now(), 0, time.Now())
	if err != nil {
		fmt.Println("ERROR - ", err.Error())
		if strings.Contains(err.Error(), "UNIQUE constraint failed:") {
			// Our unique requirement for email address has been violated
			return 0, models.ErrDuplicateEmail
		} else {
			return 0, err
		}
	}
	rowid, err := result.LastInsertId() // confirmed! sqlite3 driver has this functionality!
	if err != nil {
		fmt.Println("[ERROR] Error:", err.Error())
		return 0, err
	}
	// id has type int64; convert to int type before returning
	return int(rowid), nil
}

// list players
func (m *PlayerModel) List(status string) ([]*models.Player, error) {
	stmt := `SELECT rowid, screenName, emailAddress, loggedIn, inBattle, created, lastLogin FROM Players`
	if status == "loggedIn" {
		stmt += " WHERE loggedIn = 1"
	}
	rows, err := m.DB.Query(stmt)
	if err != nil {
		fmt.Println("[ERROR] Error:", err.Error())
		return nil, err
	}
	defer rows.Close()

	players := []*models.Player{}

	for rows.Next() {
		s := &models.Player{}
		err = rows.Scan(&s.ID, &s.ScreenName, &s.EmailAddress, &s.LoggedIn, &s.InBattle, &s.Created, &s.LastLogin)
		if err != nil {
			fmt.Println("[ERROR] Error:", err.Error())
			return nil, err
		}
		players = append(players, s)
	}
	if err = rows.Err(); err != nil {
		//fmt.Println("[ERROR] Error:", err.Error())
		return nil, err
	}

	return players, nil
}

// update player
func (m *PlayerModel) Update(id int, emailAddress string) (int, error) {
	stmt := `UPDATE Players SET emailAddress = ?, lastLogin = ? WHERE rowid = ?`
	_, err := m.DB.Exec(stmt, emailAddress, time.Now(), id)
	if err != nil {
		//fmt.Println("[ERROR] Error:", err.Error())
		if xerrors.Is(err, sql.ErrNoRows) {
			return id, models.ErrNoRecord
		}
	}
	return id, err
}

// update login
func (m *PlayerModel) UpdateLogin(id int, loggedIn bool) (int, error) {
	stmt := `UPDATE Players SET lastLogin = ?, loggedIn = ? WHERE rowid = ?`
	_, err := m.DB.Exec(stmt, time.Now(), loggedIn, id)
	if err != nil {
		//fmt.Println("[ERROR] Error encountered:", err.Error())
		if xerrors.Is(err, sql.ErrNoRows) {
			return id, models.ErrNoRecord
		}
	}
	return id, err
}