package sqlite3

import (
	"github.com/519seven/cs610/battleship/pkg/models"
	"database/sql"
)

type PositionModel struct {
	DB *sql.DB
}

func (m *PositionModel) Insert(boardID int, battleshipID int, playerID int, coordX int, coordY int, pinColor string) (int, error) {
	return 0, nil
}
func (m *PositionModel) Get(id int) (*models.Positions, error) {
	return nil, nil
}
func (m *PositionModel) List() ([]*models.Positions, error) {
	return nil, nil
}
func (m *PositionModel) Update(playerID int, battleID int, boardID int, coordX int, coordY string) (int, error) {
	pinColor := 0
	positionID := -1							// -1 means that a challenger has tried this one
	pOne := 0
	pTwo := 0
	turn := 0

	// If the opponent has a ship there, then the pinColor will be 1
	stmt := `SELECT rowid FROM Positions WHERE boardID = ? AND coordX = ? AND coordY = ?`
	err := m.DB.QueryRow(stmt, coordX, coordY).Scan(&positionID)
	if err != nil {
		return 0, err
	}
	if positionID != 0 {
		// We have a match!  This means it's a strike
		pinColor = 1
	}
	stmt = `SELECT player1ID, player2ID FROM Battles WHERE rowid = ? AND (player1ID = ? OR player2ID = ?)`
	err = m.DB.QueryRow(stmt, boardID, playerID, playerID).Scan(&pOne, &pTwo)
	// Find out which player this is
	if pOne == playerID {
		// Then player1 just moved.  We'll want to update Turn to be player2
		turn = pTwo
	} else {
		turn = pOne
	}
	stmt = `UPDATE Positions SET pinColor = ? WHERE rowid = ? AND coordX = ? AND coordY = ?; UPDATE Battles SET turn = ? WHERE rowid = ? AND (player1ID = ? OR player2ID = ?)`
	m.DB.Exec(stmt, pinColor, boardID, coordX, coordY, turn, battleID, playerID, playerID)
	return 0, nil
}
