package sqlite3

import (
	"database/sql"
	"fmt"

	"github.com/519seven/cs610/battleship/pkg/models"
)

type BattleModel struct {
	DB *sql.DB
}

func (m *BattleModel) Create(player1 int, player1Accepted bool, player2 int, player2Accepted bool) (int, error) {
	fmt.Println("Creating new battle...")
	stmt := `INSERT INTO Battles (player1, player1Accepted, player2, player2Accepted) VALUES (?, ?, ?, ?)`
	result, err := m.DB.Exec(stmt, player1, player1Accepted, player2, player2Accepted)
	if err != nil {
		return 0, err
	}
	battleID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(battleID), nil
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