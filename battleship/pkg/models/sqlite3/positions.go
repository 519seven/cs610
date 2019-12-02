package sqlite3

import (
	"database/sql"
	"fmt"
	"golang.org/x/xerrors"

	"github.com/519seven/cs610/battleship/pkg/models"
)

type PositionModel struct {
	DB *sql.DB
}

func (m *PositionModel) Insert(boardID int, battleshipID int, playerID int, coordX int, coordY int, pinColor string) (int, error) {
	return 0, nil
}
func (m *PositionModel) Get(id int) (*models.Position, error) {
	return nil, nil
}
func (m *PositionModel) List(boardID, playerID int) ([]*models.Position, error) {
	// Get coordinates that are attempted or successful strikes (gray or red) for this battle
	stmt := `SELECT
				coordX, coordY, pinColor 
			FROM 
				Positions p 
			WHERE 
				(pinColor == 'gray' OR pinColor == 'red')
			AND
				boardID = ?;`
			//AND													// playerID not necessary when querying for positions
			//	playerID = ?;`
	rows, err := m.DB.Query(stmt, boardID, playerID)
	if err != nil {
		//fmt.Println("SQL failure:", stmt, err.Error())			// debug
		return nil, err
	}
	defer rows.Close()

	positions := []*models.Position{}

	for rows.Next() {
		p := &models.Position{}
		err = rows.Scan(&p.CoordX, &p.CoordY, &p.PinColor)
		if err != nil {
			return nil, err
		}
		//fmt.Println("Position struct:", p)
		positions = append(positions, p)
		//fmt.Println("positions:", positions)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return positions, nil
}


// Update a pinColor based on player, battle, board, and coordinates - return the pinColor
func (m *PositionModel) Update(playerID int, battleID int, boardID int, coordX int, coordY string) (string, error) {
	pinColor := ""
	positionID := -1		   // -1 means that a challenger has tried this one
	pOne := 0
	pTwo := 0
	turn := 0

	// Once a strike has been logged, update whose turn it is
	stmt := `SELECT player1ID, player2ID FROM Battles WHERE rowid = ? AND (player1ID = ? OR player2ID = ?);`
	err := m.DB.QueryRow(stmt, boardID, playerID, playerID).Scan(&pOne, &pTwo)
	// Find out which player this is
	if pOne == playerID {
		// Then player1 just moved.  We'll want to update Turn to be player2
		turn = pTwo
	} else {
		turn = pOne
	}

	// Check the positions
	// - If a ship is there, update the pinColor to "red"
	// - If a ship is not there, insert a gray pin at those coordinates
	stmt = `SELECT rowid FROM Positions WHERE boardID = ? AND coordX = ? AND coordY = ?;`
	err = m.DB.QueryRow(stmt, boardID, coordX, coordY).Scan(&positionID)
	if err != nil {
		if xerrors.Is(err, sql.ErrNoRows) {
			// If we don't have a hit then add this position w/ a gray pinColor
			pinColor = "gray"
			stmt = `INSERT INTO Positions (playerID, boardID, coordX, coordY, pinColor) VALUES (?, ?, ?, ?, ?);
					UPDATE Battles SET turn = ? WHERE rowid = ? AND (player1ID = ? OR player2ID = ?);`
			m.DB.Exec(stmt, playerID, boardID, coordX, coordY, pinColor, turn, battleID, playerID, playerID)
		} else {
			// This is a sql error; bail
			fmt.Println("Error when querying for existing position:", stmt)
			return "", err	
		}
	} else {
		// Err is not nil and rows are not empty - we have a match!
		// - This means it's a strike
		pinColor = "red"
		stmt = `UPDATE Positions SET pinColor = ? WHERE boardID = ? AND coordX = ? AND coordY = ?;
				UPDATE Battles SET turn = ? WHERE rowid = ? AND (player1ID = ? OR player2ID = ?);`
		m.DB.Exec(stmt, pinColor, boardID, coordX, coordY, turn, battleID, playerID, playerID)
	}

	return pinColor, nil
}
