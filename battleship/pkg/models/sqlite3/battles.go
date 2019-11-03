package sqlite3

import (
	"database/sql"
	"fmt"

	"github.com/519seven/cs610/battleship/pkg/models"
)

type BattleModel struct {
	DB *sql.DB
}

// Create a new Battle - record the challenger (player1) and the challengee (player2)
func (m *BattleModel) Create(player1ID, player1BoardID, player2ID int) (int, error) {
	rowid := 0
	battleID := 0
	
	fmt.Println("Currently, only one game per challenger/challengee pair is supported at a time.")
	fmt.Println("Checking to see if there is already a challenge out there...")
	stmt := `SELECT rowid FROM Battles WHERE player1ID = ? AND player2ID = ? LIMIT 0, 1`
	rows, err := m.DB.Query(stmt, player1ID, player2ID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rowid)
		if err != nil {
			fmt.Println("ERROR - While retrieving battleID:", err)
		} else {
			fmt.Println("INFO - Found a pre-existing battle for this challenger/challengee pair...")
			battleID = int(rowid)
		}
	}
	if battleID > 0 {
		_, err := m.DB.Exec(`UPDATE Battles SET player1BoardID = ? WHERE player1ID = ? AND rowid = ?`, player1BoardID, player1ID, battleID)
		if err != nil {
			return 0, err
		}
		fmt.Println("INFO - Battle has been updated with fresh information...")
		return int(battleID), nil
	} else {
		fmt.Println("Battle between these two players was not found.")
		fmt.Println("Creating new battle...")
		stmt = `INSERT INTO Battles (player1ID, player1Accepted, player1BoardID, player2ID, player2Accepted) VALUES (?, ?, ?, ?, ?)`
		result, err := m.DB.Exec(stmt, player1ID, 1, player1BoardID, player2ID, 0)
		if err != nil {
			return 0, err
		}
		battleID, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(battleID), nil
	}
}

// Get - return a single battle; this is for the battle board
func (m *BattleModel) Get(rowid, battleID int) (*models.Battle, error) {
	b := &models.Battle{}

	// Get a list of battles that are available for this user
	stmt := `SELECT 
				p1.rowid as Player1ID, p1.screenName as Player1ScreenName, 
				p2.rowid as Player2ID, p2.screenName as Player2ScreenName 
				FROM Battles as b
				JOIN Players as p1 ON p1.rowid = b.player1ID
				JOIN Players as P2 ON p2.rowid = b.player2ID
				WHERE b.player2ID = ?`
	err := m.DB.QueryRow(stmt, rowid).Scan(&b.ID, &b.Player1ID, &b.Player1ScreenName, &b.Player2ID, &b.Player2ScreenName)
	if err != nil {
		fmt.Println("[ERROR] stmt", stmt, err.Error())
		return nil, err
	}
	return b, nil
}

// GetChallenger - See if there are any challengers out there
func (m *BattleModel) GetChallenger(currentUserID int) (int, error) {
	var challenger int
	stmt := `SELECT player1ID 
				FROM Battles 
				WHERE player1Accepted == true 
				AND player2ID = ? 
				AND player2Accepted == false 
				LIMIT 0, 1`
	rows, err := m.DB.Query(stmt, currentUserID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&challenger)
		if err != nil {
			fmt.Println("ERROR - While retrieving battleID:", err)
		}
	}

	return challenger, err
}

// GetAll - Get all active battles or challenges (this could be combined with GetOpen below)
func (m *BattleModel) GetChallenges(rowid int) ([]*models.Battle, error) {
	stmt := `SELECT
				b1.rowid, b2.boardName as BoardTitle, 
				p1.rowid as Player1ID, p1.screenName as Player1ScreenName, 
				p2.rowid as Player2ID, p2.screenName as Player2ScreenName 
				FROM Battles as b1
				JOIN Players as p1 ON p1.rowid = b1.player1ID
				JOIN Players as P2 ON p2.rowid = b1.player2ID
				JOIN Boards as b2 ON b2.rowid = b1.player1BoardID
				JOIN Boards as b3 ON b3.rowid = b1.player2BoardID
				WHERE b1.player2ID = ? OR b1.player1ID = ?`
	rows, err := m.DB.Query(stmt, rowid, rowid)
	if err != nil {
		fmt.Println("[ERROR] stmt", stmt, err.Error())
		return nil, err
	}
	defer rows.Close()

	battles := []*models.Battle{}

	for rows.Next() {
		b := &models.Battle{}
		err = rows.Scan(&b.ID, &b.BoardTitle, &b.Player1ID, &b.Player1ScreenName, &b.Player2ID, &b.Player2ScreenName)
		if err != nil {
			fmt.Println("[ERROR] Error:", err.Error())
			return nil, err
		}
		battles = append(battles, b)
	}
	if err = rows.Err(); err != nil {
		//fmt.Println("[ERROR] Error:", err.Error())
		return nil, err
	}

	return battles, nil
}

// GetOpen - Get a list of open challenges
func (m *BattleModel) GetOpen(rowid, battleID int) ([]*models.Battle, error) {
	// Get a list of battles that are available for this user
	stmt := `SELECT 
				p1.rowid as Player1ID, p1.screenName as Player1ScreenName, 
				p2.rowid as Player2ID, p2.screenName as Player2ScreenName 
				FROM Battles as b
				JOIN Players as p1 ON p1.rowid = b.player1ID
				JOIN Players as P2 ON p2.rowid = b.player2ID
				WHERE b.player2ID = ?`
	rows, err := m.DB.Query(stmt, rowid)
	if err != nil {
		fmt.Println("[ERROR] stmt", stmt, err.Error())
		return nil, err
	}
	defer rows.Close()

	battles := []*models.Battle{}

	for rows.Next() {
		b := &models.Battle{}
		err = rows.Scan(&b.ID, &b.Player1ID, &b.Player1ScreenName, &b.Player2ID, &b.Player2ScreenName)
		if err != nil {
			fmt.Println("[ERROR] Error:", err.Error())
			return nil, err
		}
		battles = append(battles, b)
	}
	if err = rows.Err(); err != nil {
		//fmt.Println("[ERROR] Error:", err.Error())
		return nil, err
	}

	return battles, nil
}


func (m *BattleModel) UpdateChallenge(player1 int, player2 int, player2Accepted bool, battleID int) (error) {
	stmt := `UPDATE Battles SET player2Accepted = ? WHERE rowid = ?`
	_, err := m.DB.Exec(stmt, player2Accepted, battleID)
	if err != nil {
		return err
	}
	return err
}

func (m *BattleModel) UpdateTurn(player1 int, player2 int, nextTurn int, battleID int) (error) {
	// Swap the value of nextTurn
	if nextTurn == player1 {
		nextTurn = player2
	} else {
		nextTurn = player1
	}
	stmt := `UPDATE Battles SET nextTurn = ? WHERE rowid = ?`
	_, err := m.DB.Exec(stmt, nextTurn, battleID)
	if err != nil {
		return err
	}
	return err
}