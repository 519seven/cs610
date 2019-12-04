package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/xerrors"

	"github.com/519seven/cs610/battleship/pkg/models"
)

type PositionModel struct {
	DB *sql.DB
}


// Check the Battles board to see if one of the players has 5 sunken ships
func (m *PositionModel) CheckWinner(playerID, battleID int) (bool, error) {
	var pOne int = 0
	var pTwo int = 0
	var numberOfShipsSunk int = 0

	stmt := `SELECT player1ID, player2ID FROM Battles WHERE rowid = ? AND (player1ID = ? OR player2ID = ?)`
	err := m.DB.QueryRow(stmt, battleID, playerID).Scan(&pOne, &pTwo)
	if err != nil {
		return false, nil
	}
	if pOne == playerID {
		// Get player1SunkenShips count
		stmt = `SELECT player1SunkenShips FROM BATTLES WHERE rowid = ? AND player1ID = ?`
		err = m.DB.QueryRow(stmt, battleID, playerID).Scan(&numberOfShipsSunk)
		if err != nil {
			return false, nil
		}
	} else if pTwo == playerID {
		// Get player2SunkenShips count
		stmt = `SELECT player2SunkenShips FROM BATTLES WHERE rowid = ? AND player2ID = ?`
		err = m.DB.QueryRow(stmt, battleID, playerID).Scan(&numberOfShipsSunk)
		if err != nil {
			return false, nil
		}
	}
	if numberOfShipsSunk == 5 {
		return true, nil
	}
	return false, nil
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


// Update positions pinColor
// - Query player, battle, board, and coordinates
// - Return the pinColor and shipType (if sunk)
// - Update whose turn it is
func (m *PositionModel) Update(playerID int, battleID int, boardID int, coordX int, coordY, secretTurn string) (string, string, bool, error) {
	pinColor := ""
	positionID := -1
	var pOne int = 0 
	var pTwo int = 0
	var turn int = 0
	var sunkenShip string = ""
	var sunkenShipSQL string = ""
	var winner bool = false

	stmt := `SELECT player1ID, player2ID FROM Battles WHERE rowid = ? AND (player1ID = ? OR player2ID = ?);`
	err := m.DB.QueryRow(stmt, battleID, playerID, playerID).Scan(&pOne, &pTwo)
	// Find out which player this is
	if pOne == playerID {
		fmt.Printf("Player1 (%d) just launched. Strike will be recorded and Player2 will go next...", pOne)
		// Then player1 just moved.  We'll want to update Turn to be player2
		sunkenShipSQL = `UPDATE Battles SET Player1SunkenShips = Player1SunkenShips +1 WHERE rowid = ?`
		turn = pTwo
	} else {
		fmt.Printf("Player2 (%d) just launched. Strike will be recorded and Player1 will go next...", pTwo)
		sunkenShipSQL = `UPDATE Battles SET Player2SunkenShips = Player2SunkenShips +1 WHERE rowid = ?`
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
					UPDATE Battles SET turn = ?, secretTurn = ? WHERE rowid = ? AND (player1ID = ? OR player2ID = ?);`
			_, err = m.DB.Exec(stmt, playerID, boardID, coordX, coordY, pinColor, turn, secretTurn, battleID, playerID, playerID)
			if err != nil {
				return "", "", false, err
			}
		} else {
			// This is a sql error; bail
			fmt.Println("Error when querying for existing position:", stmt)
			return "", "", false, err	
		}
	} else {
		// Err is not nil and rows are not empty - we have a match!
		// - This means it's a strike
		pinColor = "red"
		stmt = `UPDATE Positions SET pinColor = ? WHERE boardID = ? AND coordX = ? AND coordY = ?;
				UPDATE Battles SET turn = ?, secretTurn = ? WHERE rowid = ? AND (player1ID = ? OR player2ID = ?);`
		_, err = m.DB.Exec(stmt, pinColor, boardID, coordX, coordY, turn, secretTurn, battleID, playerID, playerID)
		if err != nil {
			return "", "", false, err
		}
		// Find out if a ship has been sunk...
		var pinCount int = 0
		stmt = `SELECT COUNT(*), s.shipType 
				FROM Positions p 
				LEFT OUTER JOIN Ships s ON s.rowid = p.shipID 
				WHERE boardID = ? AND pinColor = 'red' AND shipType IN 
				(SELECT s.shipType 
					FROM Positions p 
					LEFT OUTER JOIN Ships s ON s.rowid = p.shipid 
					WHERE p.coordX = ? AND p.coordy = ?)`
		// This SQL statement will give us a count of the number of ships that have red pins
		// - If the number of red pins that are present matches the maximum number for that ship
		//   then, the ship has been sunk and we return the ship name
		err := m.DB.QueryRow(stmt, boardID, coordX, coordY).Scan(&pinCount, &sunkenShip)
		if err != nil {
			return "", "", false, err
		}
		switch sunkenShip {
		case "battleship":
			if pinCount != 4 { sunkenShip = "" }
		case "carrier":
			if pinCount != 5 { sunkenShip = "" }
		case "cruiser":
			if pinCount != 3 { sunkenShip = "" }
		case "destroyer":
			if pinCount != 2 { sunkenShip = "" }
		case "submarine":
			if pinCount != 3 { sunkenShip = "" }
		default:
			fmt.Println("We got a row from the database but none of the shipTypes matched???")
		}
		if pinCount > 0 && sunkenShip != "" {
			// Update this player's sunken ship counter
			_, err := m.DB.Exec(sunkenShipSQL, battleID)
			if err != nil {
				return pinColor, sunkenShip, winner, errors.New("Unable to update sunkenShip counter")
			}
			winner, err := m.CheckWinner(playerID, battleID)
			if err != nil {
				return pinColor, sunkenShip, winner, errors.New("Error while checking winner")
			}
		}
	}
	return pinColor, sunkenShip, winner, nil
}
