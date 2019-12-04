package models

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrMissingBoardID = errors.New("Missing BoardID")
	ErrNoRecord = errors.New("models: no matching record found")
	ErrInvalidCredentials = errors.New("models: invalid credentials")
	ErrDuplicateEmail = errors.New("models: duplicate email")
)

type About struct {
	AuthenticatedUserID		int
	ScreenName				string
}
type Battle struct {
	ID        				int
	AuthenticatedPlayerID	int
	Title					string
	Player1ID 				int
	Player1ScreenName		string
	Player1BoardID			int
	Player1Accepted			bool
	Player2ID 				int
	Player2ScreenName		string
	Player2BoardID			int
	Player2Accepted			bool
	ChallengerBoardName		sql.NullString
	ChallengeDate			time.Time
	Turn      				sql.NullInt64
}

type Board struct {
	ID						int
	Title   				string
	PlayerID 				int
	BattleID  				sql.NullInt64
	Created 				time.Time
}

type Login struct {
	ID      	   			int
	ScreenName 				string
	IsActive   				sql.NullString
	LastLogin  				time.Time
}

type Player struct {
	ID        	 			int
	EmailAddress			string
	HashedPassword			string
	ScreenName 				string
	LoggedIn   				sql.NullString
	InBattle   				sql.NullString
	Created  				sql.NullString
	LastLogin  				sql.NullString
}

type Position struct {
	ID      				int
	PlayerID 				int
	BattleID  				sql.NullInt64
	PositionID      		int
	ShipType		 		sql.NullString
	CoordX					int
	CoordY					string
	PinColor				string
}

type Signup struct {
	ID       	  			int
	ScreenName 				string
	IsActive   				sql.NullString
	LastLogin 	 			time.Time
}

type Ship struct {
	ID     					int
	Length 					int
	Title  					string
}
