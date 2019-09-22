package sqlite3

import (
	"519seven/battleship/pkg/models"
	"database/sql"
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
func (m *PositionModel) List() ([]*models.Position, error) {
	return nil, nil
}
