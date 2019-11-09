package sqlite3

import (
	"database/sql"
	"fmt"
	"golang.org/x/xerrors"
	"strings"

	"github.com/519seven/cs610/battleship/pkg/models"
)

type BoardModel struct {
	DB *sql.DB
}

// trim last char from end of string
func trimSuffix(s, suffix string) string {
    if strings.HasSuffix(s, suffix) {
        s = s[:len(s)-len(suffix)]
    }
    return s
}

// return the next alphabet character but stop at maxCharacters
func GetNextChar(character string, maxCharacters uint) byte {
	var alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := range alphabet {
		if string(alphabet[i]) == character {
			if i+1 < int(maxCharacters) {
				return alphabet[i+1]
			} else {
				return alphabet[i]
			}
		}
	}
	return 'z'
}
func MatchFound(coordinates string, stringOfCoords string) bool {
	//fmt.Println("Looking for ", coordinates)
	if strings.Contains(stringOfCoords, coordinates) {
		return true
	}
	return false
}

// Choose Board - I can't store the boardID in my session so...
// 				  update the database and keep one "selected" board at any given point in time
func (m *BoardModel) ChooseBoard(rowid, boardID int, action string) (int, error) {
	if action == "update" {
		stmt := `UPDATE Boards SET isChosen = 0;`
		_, err := m.DB.Exec(stmt)
		if err != nil {
			return 0, err
		}
		stmt = `UPDATE Boards SET isChosen = 1 WHERE rowid = ?`
		_, err = m.DB.Exec(stmt, boardID)
		if err != nil {
			return 0, err
		}
		return boardID, nil
	} else if action == "select" {
		var boardID int
		stmt := `SELECT rowid FROM Boards WHERE isChosen = 1;`
		err := m.DB.QueryRow(stmt).Scan(&boardID)
		if err != nil {
			return 0, err
		}
		return boardID, nil
	} else {
		fmt.Println("Default action for ChooseBoard (nothing is happening, check logic)")
		return boardID, nil
	}
}

// Create a board if one with the same name doesn't already exist (belonging to this user)
func (m *BoardModel) Create(rowid int, boardName string) (int, error) {
	var boardID int64
	// first check to make sure a board with the same name doesn't already exist
	stmt := `SELECT rowid FROM Boards WHERE boardName = ? AND playerID = ?`
	err := m.DB.QueryRow(stmt, boardName, rowid).Scan(boardID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		fmt.Println("[ERROR] Error encountered, returning to calling func:", err.Error())
		return 0, err
	} else if boardID > 0 {
		fmt.Println("Found existing board, returning its id")
		return int(boardID), nil
	}
	fmt.Println("Creating new board...")
	stmt = `INSERT INTO Boards (boardName, playerID) VALUES (?, ?)`
	result, err := m.DB.Exec(stmt, boardName, rowid)
	if err != nil {
		return 0, err
	}
	boardID, err = result.LastInsertId() // confirmed! sqlite3 driver has this functionality!
	if err != nil {
		return 0, err
	}
	return int(boardID), nil
}

// Get board info - name of board
func (m *BoardModel) GetInfo(playerID, boardID int) (*models.Board, error) {
	stmt := `SELECT 
		b.rowid as ID, b.boardName as Title, b.playerID as playerID, b.created
		FROM Boards b
		WHERE b.rowid = ? AND b.playerID = ?`
	b := &models.Board{}

	err := m.DB.QueryRow(stmt, boardID, playerID).Scan(
			&b.ID, &b.Title, &b.PlayerID, &b.Created)
	if err != nil {
		if xerrors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return b, nil
}

// Get positions on board
func (m *BoardModel) GetPositions(rowid int) ([]*models.Positions, error) {
	fmt.Println("boardID = ", rowid)
	if rowid == 0 {
		return nil, models.ErrMissingBoardID
	}
	positions := []*models.Positions{}
	stmt := `SELECT 
		b.rowid as boardID, p.playerID as playerID, 
		p.rowid as positionID, s.shipType, p.coordX, p.coordY, p.pinColor
		FROM Boards b
		LEFT OUTER JOIN Positions p ON
		p.boardID = b.rowid 
		LEFT OUTER JOIN Ships s ON
		s.rowid = p.shipID
		WHERE b.rowid = ?` // and playerID = this user's ID
	fmt.Println("board positions sql for ", rowid, ">>", stmt)
	rows, err := m.DB.Query(stmt, rowid)
	for rows.Next() {
		p := &models.Positions{}
		// Assign fields in rowset to Board model's "properties"
		err = rows.Scan(&p.ID, &p.PlayerID, &p.PositionID, &p.ShipType, &p.CoordX, &p.CoordY, &p.PinColor)
		if err != nil {
			return nil, err
		}
		positions = append(positions, p)
	}
	if err != nil {
		if xerrors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}

	return positions, nil
}

// Insert coordinates for a board
func (m *BoardModel) Insert(playerID int, boardID int, shipName string, arrayOfCoords []string) (int, error) {
	// Get shipID
	var shipID int
	stmt := "SELECT rowid FROM Ships WHERE shipType = ? LIMIT 1"
	rows, err := m.DB.Query(stmt, shipName)
	if err != nil {
		// Can't move forward without a shipID...Not sure we want to guess what it could be, either
		// log.Fatal()!!!
		fmt.Println("[ERROR] retrieving the shipID: ", err)
		// I guess we can recreate the Ships table but it's strange it's not there already
		// "import cycle" when importing initializeDB
		/*
		db, err := initializeDB(*dsn, *initdb)
		if err != nil {
			errorLog.Fatal(err)
		}
		defer db.Close()	 
		*/
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&shipID)
		if err != nil {
			//log.Fatal()
			fmt.Println("ERROR - While retrieving shipID:", err)
		}
	}
	err = rows.Err()
	if err != nil {
		fmt.Println("[FATAL]:", err)
	}
	// Save the coordinates and return to calling function
	for _, rc := range arrayOfCoords {
		s := strings.Split(rc, ",")
		//fmt.Println(s)
		row, col := s[0], s[1]
		colStr := string(col)
		stmt := "INSERT INTO Positions (boardID, shipID, playerID, coordX, coordY, pinColor) VALUES (?, ?, ?, ?, ?, ?)"
		_, err := m.DB.Exec(stmt, boardID, shipID, playerID, row, colStr, 0)
		if err != nil {
			fmt.Println("[ERROR] inserting position: ", err, boardID, shipName, playerID, row, colStr)
		}
		//fmt.Println("[INFO] inserted positions into board with id=", boardID)
	}
	// Return
	fmt.Println("Done with", shipName, "\nReturning control back to createBoard")
	return 0, nil
}

func (m *BoardModel) List(rowid int) ([]*models.Board, error) {
	stmt := `SELECT rowid, boardName, playerID, created FROM Boards 
	WHERE playerID = ?
	ORDER BY created DESC LIMIT 10`

	rows, err := m.DB.Query(stmt, rowid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	boards := []*models.Board{}

	for rows.Next() {
		s := &models.Board{}
		// Assign fields in rowset to Board model's "properties"
		err = rows.Scan(&s.ID, &s.Title, &s.PlayerID, &s.Created)
		if err != nil {
			return nil, err
		}
		boards = append(boards, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return boards, nil
}

func (m *BoardModel) Update(rowid int, boardName string, playerID int) (int, error) {
	// to split over multpile lines, use backquotes not double quotes
	stmt := `UPDATE Boards SET boardName = ?, playerID = ?, WHERE rowid = ?`
	_, err := m.DB.Exec(stmt, boardName, playerID, rowid)
	if err != nil {
		return 0, err
	}
	// rowid has type int64; convert to int type before returning
	return int(rowid), nil
}
