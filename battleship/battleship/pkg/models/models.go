package models

import (
	"database/sql"
	"errors"
	"time"
)

var ErrNoRecord = errors.New("models: no matching record found")

type Battle struct {
	ID        int
	Player1ID int
	Player2ID int
	Turn      int
}

type Board struct {
	ID      int
	Title   string
	OwnerID int
	GameID  int
	Created time.Time
}

type Player struct {
	ID         int
	ScreenName string
	IsActive   sql.NullString
	LastLogin  time.Time
}

type Position struct {
	ID           int
	BoardID      int
	BattleshipID int
}

type Ship struct {
	ID     int
	Length int
	Title  string
}
