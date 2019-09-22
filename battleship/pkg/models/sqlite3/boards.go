package sqlite3

import (
	"519seven/battleship/pkg/models"
	"database/sql"
	"errors"
)

type BoardModel struct {
	DB *sql.DB
}

//func (m *BoardModel) Display(id int) (*models.Board, error)

func (m *BoardModel) Get(id int) (*models.Board, error) {
	//	stmt := `SELECT id, boardName, userID, gameID, created FROM Boards WHERE id = ?` // and userID = this user's ID

	//	row := m.DB.QueryRow(stmt, id)

	//	s := &models.Board{}
	//	err := row.Scan(&s.ID, &s.Title, &s.OwnerID, &s.GameID, &s.Created)
	stmt := `SELECT id, boardName, userID, gameID, created FROM Boards WHERE id = ?` // and userID = this user's ID
	s := &models.Board{}
	err := m.DB.QueryRow(stmt, id).Scan(&s.ID, &s.Title, &s.OwnerID, &s.GameID, &s.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}
func (m *BoardModel) Insert(boardName string, userID int) (int, error) {
	// to split over multpile lines, use backquotes not double quotes
	stmt := `INSERT INTO Boards (boardName, userID) VALUES (?, ?)`
	result, err := m.DB.Exec(stmt, boardName, userID)
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
func (m *BoardModel) List(userID int) ([]*models.Board, error) {
	stmt := `SELECT id, boardName, userID, gameID, created FROM Boards 
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
func (m *BoardModel) Update(id int, boardName string, userID int, gameID int) (int, error) {
	// to split over multpile lines, use backquotes not double quotes
	stmt := `UPDATE Boards SET boardName = ?, userID = ?, gameID= ? WHERE id = ?`
	_, err := m.DB.Exec(stmt, boardName, userID, gameID, id)
	if err != nil {
		return 0, err
	}
	// id has type int64; convert to int type before returning
	return int(id), nil
}
