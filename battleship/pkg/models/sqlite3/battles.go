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
func (m *BattleModel) Create(player1ID int, player2ID int) (int, error) {
	var rowid int
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
			return int(rowid), nil
		}
	}
	fmt.Println("Battle between these two players was not found.")
	fmt.Println("Creating new battle...")
	stmt = `INSERT INTO Battles (player1ID, player1Accepted, player2ID, player2Accepted) VALUES (?, ?, ?, ?)`
	result, err := m.DB.Exec(stmt, player1ID, 1, player2ID, 0)
	if err != nil {
		return 0, err
	}
	battleID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(battleID), nil
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

func (m *BattleModel) Get(id int) (*models.Battle, error) {
	return nil, nil
}
func (m *BattleModel) List() ([]*models.Battle, error) {
	return nil, nil
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