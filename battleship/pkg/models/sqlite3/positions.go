package sqlite3

import (
	"github.com/519seven/cs610/battleship/pkg/models"
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
func (m *PositionModel) Update(boardID int, playerID int, coordX int, coordY string, pinColor int) (int, error) {
	return 0, nil
}
