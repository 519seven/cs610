package sqlite3

import (
	"database/sql"
	"golang.org/x/xerrors"
	"fmt"
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

// Create a board if one with the same name doesn't already exist (belonging to this user)
func (m *BoardModel) Create(boardName string) (int, error) {
	var boardID int64
	userID := 1 												// get from secure location
	// first check to make sure a board with the same name doesn't already exist
	stmt := `SELECT rowid FROM BOARDS WHERE boardName = ? AND userID = ?`
	err := m.DB.QueryRow(stmt, boardName, userID).Scan(boardID)
	if err != nil && err.Error() != "sql: no rows in result set" {
		fmt.Println("[ERROR] Error encountered, returning to calling func:", err.Error())
		return 0, err
	} else if boardID > 0 {
		fmt.Println("Found existing board, returning its id")
		return int(boardID), nil
	}
	fmt.Println("Creating new board...")
	stmt = `INSERT INTO Boards (boardName, userID) VALUES (?, ?)`
	result, err := m.DB.Exec(stmt, boardName, userID)
	if err != nil {
		return 0, err
	}
	boardID, err = result.LastInsertId() // confirmed! sqlite3 driver has this functionality!
	if err != nil {
		return 0, err
	}
	return int(boardID), nil
}

// Get board info - name of board and the positions that have been saved on it
func (m *BoardModel) Get(rowid int) (*models.PositionsOnBoard, error) {
	stmt := `SELECT b.rowid as boardID, boardName, b.userID, gameID, created,
		p.rowid as positionID, s.shipType, p.userID, p.coordX, p.coordY, p.pinColor
		FROM Boards b
		LEFT OUTER JOIN Positions p ON
		p.boardID = b.rowid 
		LEFT OUTER JOIN Ships s ON
		s.rowid = p.shipID
		WHERE b.rowid = ?` // and userID = this user's ID
	p := &models.PositionsOnBoard{}
	err := m.DB.QueryRow(stmt, rowid).Scan(&p.BoardID, &p.BoardName, &p.OwnerID, &p.GameID, &p.Created,
		&p.PositionID, &p.ShipType, &p.UserID, &p.CoordX, &p.CoordY, &p.PinColor)
	if err != nil {
		if xerrors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return p, nil
}

// Insert coordinates for a board
func (m *BoardModel) Insert(boardID int, shipName string, arrayOfCoords []string) (int, error) {
	// get userID from a trusted location
	userID := 1
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
			fmt.Println("[ERROR] retrieving shipID.  Unable to continue.  Error msg: ", err)
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
		stmt := "INSERT INTO Positions (boardID, shipID, userID, coordX, coordY, pinColor) VALUES (?, ?, ?, ?, ?, ?)"
		_, err := m.DB.Exec(stmt, boardID, shipID, userID, row, colStr, 0)
		if err != nil {
			fmt.Println("[ERROR] inserting position: ", err, boardID, shipName, userID, row, colStr)
		}
		//fmt.Println("[INFO] inserted positions into board with id=", boardID)
	}
	// Return
	fmt.Println("Done with", shipName, "\nReturning control back to createBoard")
	return 0, nil
}

func (m *BoardModel) List(userID int) ([]*models.Board, error) {
	stmt := `SELECT rowid, boardName, userID, gameID, created FROM Boards 
	WHERE userID = ?
	ORDER BY created DESC LIMIT 10`

	rows, err := m.DB.Query(stmt, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	boards := []*models.Board{}

	for rows.Next() {
		s := &models.Board{}
		// Assign fields in rowset to Board model's "properties"
		err = rows.Scan(&s.ID, &s.Title, &s.OwnerID, &s.GameID, &s.Created)
		if err != nil {
			return nil, err
		}
		boards = append(boards, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return boards, nil
	/*
		s, err := app.boards.List()
		if err != nil {
			app.serverError(w, err)
			return
		}

		for _, board := range s {
			fmt.Fprintf(w, "%v\n", board)
		}
	*/
}

func (m *BoardModel) Update(rowid int, boardName string, userID int, gameID int) (int, error) {
	// to split over multpile lines, use backquotes not double quotes
	stmt := `UPDATE Boards SET boardName = ?, userID = ?, gameID= ? WHERE rowid = ?`
	_, err := m.DB.Exec(stmt, boardName, userID, gameID, rowid)
	if err != nil {
		return 0, err
	}
	// rowid has type int64; convert to int type before returning
	return int(rowid), nil
}
