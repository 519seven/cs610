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
	OwnerID 			int
	GameID  			sql.NullInt64
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

type Position struct {
	ID           		int
	BoardID      		int
	ShipID		 		int
	UserID				int
	CoordX				int
	CoordY				string
	PinColor			int
}

type PositionsOnBoard struct {
	BoardID      		int
	BoardName   		string
	OwnerID 			int
	GameID  			sql.NullInt64
	Created 			time.Time
	PositionID      	int
	ShipType	 		string
	UserID				int
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
