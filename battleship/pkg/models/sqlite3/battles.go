package sqlite3

import (
	"519seven/battleship/pkg/models"
	"database/sql"
)

type BattleModel struct {
	DB *sql.DB
}

func (m *BattleModel) Insert(player1ID int, player2ID int, turn int) (int, error) {
	return 0, nil
}
func (m *BattleModel) Get(id int) (*models.Battle, error) {
	return nil, nil
}
func (m *BattleModel) List() ([]*models.Battle, error) {
	return nil, nil
}
