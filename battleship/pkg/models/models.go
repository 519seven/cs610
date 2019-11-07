package models

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNoRecord = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail = errors.New("models: duplicate email")
)

type Battle struct {
	ID        			int
	AuthenticatedUserID	int
	BoardTitle			string
	Player1ID 			int
	Player1ScreenName	string
	Player1Accepted		bool
	Player2ID 			int
	Player2ScreenName	string
	Player2Accepted		bool
	ChallengeDate		time.Time
	Turn      			sql.NullInt64
}

type Board struct {
	ID      			int
	Title   			string
	PlayerID 			int
	BattleID  			sql.NullInt64
	Created 			time.Time
}

type Login struct {
	ID         			int
	ScreenName 			string
	IsActive   			sql.NullString
	LastLogin  			time.Time
}

type Player struct {
	ID         			int
	EmailAddress		string
	HashedPassword		string
	ScreenName 			string
	LoggedIn   			sql.NullString
	InBattle   			sql.NullString
	Created  			time.Time
	LastLogin  			time.Time
}

type Positions struct {
	ID      			int
	PlayerID 			int
	BattleID  			sql.NullInt64
	PositionID      	int
	ShipType	 		string
	CoordX				int
	CoordY				string
	PinColor			int
}

type Signup struct {
	ID         			int
	ScreenName 			string
	IsActive   			sql.NullString
	LastLogin  			time.Time
}

type Ship struct {
	ID     				int
	Length 				int
	Title  				string
}
