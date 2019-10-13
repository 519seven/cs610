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
	GameID  sql.NullInt64
	Created time.Time
}

type Player struct {
	ID         int
	ScreenName string
	IsActive   sql.NullString
	LastLogin  time.Time
}

type Position struct {
	ID           	int
	BoardID      	int
	ShipID		 	int
	UserID			int
	CoordX			int
	CoordY			string
	PinColor		int
}

type Ship struct {
	ID     int
	Length int
	Title  string
}

type PositionsOnBoard struct {
	BoardID      	int
	BoardName   	string
	OwnerID 		int
	GameID  		sql.NullInt64
	Created 		time.Time
	PositionID      int
	ShipType	 	string
	UserID			int
	CoordX			int
	CoordY			string
	PinColor		int
}
