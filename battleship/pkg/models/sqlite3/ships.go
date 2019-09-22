package sqlite3

import (
	"github.com/519seven/cs610/battleship/pkg/models"
	"database/sql"
)

type ShipModel struct {
	DB *sql.DB
}

func (m *ShipModel) Insert(shipType string, shipLength int) (int, error) {
	return 0, nil
}
func (m *ShipModel) Get(id int) (*models.Ship, error) {
	return nil, nil
}
func (m *ShipModel) List() ([]*models.Ship, error) {
	return nil, nil
}
