package sqlite3

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/519seven/cs610/battleship/pkg/models"
)

type BattleModel struct {
	DB *sql.DB
}

// Accept a challenge (battle)
func (m *BattleModel) Accept(player2ID, boardID, battleID int) (int, error) {
	var player2IDFromDB int; player2IDFromDB = 0
	// Check to be sure that the person accepting this battle is matches the "player2ID"
	stmt := `SELECT player2ID FROM Battles WHERE rowid = ? AND player2Accepted = false`
	err := m.DB.QueryRow(stmt, battleID).Scan(&player2IDFromDB)
	if err != nil {
		return 0, err
	}

	if player2ID == player2IDFromDB {
		// Only player2 can accept a challenge
		stmt = `UPDATE Battles SET player2Accepted = true, player2BoardID = ? WHERE player2ID = ? AND rowid = ?`
		_, err := m.DB.Exec(stmt, boardID, player2ID, battleID)
		if err != nil {
			return 0, err
		}
		return battleID, nil
	}
	return 0, errors.New("Player mistmatch")
}


// GetBoardOwner - Ensure the player is the owner of this board and this board is part of this battle
func (m *BattleModel) CheckBoardOwner(playerID, battleID, boardID int) bool {
	var battleIDEntry int; battleIDEntry = 0
	stmt := `SELECT rowid
				FROM Battles
				WHERE
				((player1ID = ? AND player1BoardID = ?)
				OR
				(player2ID = ? AND player2BoardID = ?))
				AND rowid = ?`
	//fmt.Println("CheckBoardOwner:", stmt, ":playerID:", playerID, ":battleID:", battleID, ":boardID:", boardID)			// debug
	err := m.DB.QueryRow(stmt, playerID, boardID, playerID, boardID, battleID).Scan(&battleIDEntry)
	fmt.Println("battleIDEntry:", battleIDEntry)
	if err != nil {
		//fmt.Println("Error in CheckBoardOwner:", err.Error())					// debug
		return false
	}

	if battleIDEntry != 0 {
		//fmt.Println("board owner is confirmed...")							// debug
		return true
	}
	return false
}


// Check that the person logged in should be updating this battle
func (m *BattleModel) CheckChallenger(player1ID, battleID, player2BoardID int) bool {
	return true	// Return true
}


// Create a new Battle - record the challenger (player1) and the challengee (player2)
func (m *BattleModel) Create(player1ID, player1BoardID, player2ID int, secretTurn string) (int, error) {
	var rowid int; rowid = 0
	var battleID int; battleID = 0
	
	fmt.Println("Currently, only one game per challenger/challengee pair is supported at a time.")
	fmt.Println("Checking to see if there is already a challenge out there...")
	stmt := `SELECT rowid FROM Battles WHERE player1ID = ? AND player2ID = ? LIMIT 0, 1`
	rows, err := m.DB.Query(stmt, player1ID, player2ID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&rowid)
		if err != nil {
			fmt.Println("ERROR - While retrieving battleID:", err)
		} else {
			fmt.Println("INFO - Found a pre-existing battle for this challenger/challengee pair...")
			battleID = int(rowid)
		}
	}
	if battleID > 0 {
		stmt = `UPDATE Battles SET player1BoardID = ? WHERE player1ID = ? AND rowid = ?`
		_, err := m.DB.Exec(stmt, player1BoardID, player1ID, battleID)
		if err != nil {
			return 0, err
		}
		fmt.Println("INFO - Battle has been updated with fresh information...")
		return int(battleID), nil
	} else {
		fmt.Println("Battle between these two players was not found.")
		fmt.Println("Creating new battle...")
		stmt = `INSERT INTO Battles (player1ID, player1Accepted, player1BoardID, player2ID, player2Accepted, turn, secretTurn) VALUES (?, ?, ?, ?, ?, ?, ?)`
		// The opponent will always go first
		if err != nil {
			return 0, err
		}	
		result, err := m.DB.Exec(stmt, player1ID, 1, player1BoardID, player2ID, 0, player2ID, secretTurn)
		if err != nil {
			return 0, err
		}
		battleID, err := result.LastInsertId()
		if err != nil {
			return 0, err
		}
		return int(battleID), nil
	}
}

// Get - return a single battle; this is for the battle board
func (m *BattleModel) Get(playerID, battleID int) (*models.Battle, error) {
	b := &models.Battle{}

	// Get a single battle that is available for this user
	stmt := `SELECT b.rowid,
				p2.screenName||' vs. '||p1.screenName as battleTitle, 
				p1.rowid as Player1ID, p1.screenName as Player1ScreenName, IFNULL(b.player1BoardID, 0),
				p2.rowid as Player2ID, p2.screenName as Player2ScreenName, IFNULL(b.player2BoardID, 0)
				FROM Battles as b
				JOIN Players as p1 ON p1.rowid = b.player1ID
				JOIN Players as P2 ON p2.rowid = b.player2ID
				WHERE b.player1ID = ? OR b.player2ID = ? AND b.rowid = ?`
	fmt.Println("battles get stmt:", stmt)
	err := m.DB.QueryRow(stmt, playerID, playerID, battleID).Scan(
		&b.ID, &b.Title, 
		&b.Player1ID, &b.Player1ScreenName, &b.Player1BoardID, 
		&b.Player2ID, &b.Player2ScreenName, &b.Player2BoardID)
	if err != nil {
		return nil, err
	}
	return b, nil
}


// GetChallenger - See if there are any challengers out there
func (m *BattleModel) GetChallenger(currentPlayerID int) (int, error) {
	var challenger int
	stmt := `SELECT player1ID 
				FROM Battles 
				WHERE player1Accepted == true 
				AND player2ID = ? 
				AND player2Accepted == false 
				LIMIT 0, 1`
	rows, err := m.DB.Query(stmt, currentPlayerID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&challenger)
		if err != nil {
			return 0, err
		}
	}

	return challenger, err
}


// GetAll - Get all active battles or challenges (this could be combined with GetOpen below)
func (m *BattleModel) GetChallenges(rowid int) ([]*models.Battle, error) {

	// I think I want to swap player1 and player2 things in the second SELECT ***
	
	stmt := `
	SELECT 
	b1.rowid, player1ID, p1.screenName as challenger, boardName, 
	player1Accepted, player2ID, p2.screenName as opponent, 
	player2Accepted, challengeDate as dateAsked, turn 
	FROM Battles b1 
	LEFT OUTER JOIN Boards bo1 ON bo1.rowid = b1.player1BoardID 
	LEFT OUTER JOIN Players p1 ON p1.rowid = b1.player1ID 
	LEFT OUTER JOIN Players p2 ON p2.rowid = b1.player2ID 
	WHERE b1.player1ID = ? 
	UNION 
	SELECT
	b2.rowid, player1ID, p4.screenName as challenger, boardName,
	player1Accepted, player2ID, p3.screenName as opponent, 
	player2Accepted, challengeDate as dateAsked, turn 
	FROM Battles b2 
	LEFT OUTER JOIN Boards bo2 ON bo2.rowid = b2.player2BoardID 
	LEFT OUTER JOIN Players p3 ON p3.rowid = b2.player2ID 
	LEFT OUTER JOIN Players p4 ON p4.rowid = b2.player1ID 
	WHERE b2.player2ID = ?;
	`
	fmt.Println("The big sql stmt:", stmt)
	rows, err := m.DB.Query(stmt, rowid, rowid)
	if err != nil {
		fmt.Println("The big sql stmt:", stmt, err.Error())
		return nil, err
	}
	defer rows.Close()

	battles := []*models.Battle{}

	for rows.Next() {
		b := &models.Battle{}
		err = rows.Scan(
			&b.ID, 
			&b.Player1ID, &b.Player1ScreenName, &b.ChallengerBoardName, 
			&b.Player1Accepted, &b.Player2ID, &b.Player2ScreenName, 
			&b.Player2Accepted, &b.ChallengeDate, &b.Turn)
		if err != nil {
			return nil, err
		}
		battles = append(battles, b)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return battles, nil
}

// GetOpen - Get a list of open challenges
func (m *BattleModel) GetOpen(rowid, battleID int) ([]*models.Battle, error) {
	// Get a list of battles that are available for this user
	stmt := `SELECT 
				p1.rowid as Player1ID, p1.screenName as Player1ScreenName, 
				p2.rowid as Player2ID, p2.screenName as Player2ScreenName 
				FROM Battles as b
				JOIN Players as p1 ON p1.rowid = b.player1ID
				JOIN Players as P2 ON p2.rowid = b.player2ID
				WHERE b.player2ID = ?`
	rows, err := m.DB.Query(stmt, rowid)
	if err != nil {
		fmt.Println("This is the sql statement:", stmt)
		return nil, err
	}
	defer rows.Close()

	battles := []*models.Battle{}

	for rows.Next() {
		b := &models.Battle{}
		err = rows.Scan(&b.ID, &b.Player1ID, &b.Player1ScreenName, &b.Player2ID, &b.Player2ScreenName)
		if err != nil {
			return nil, err
		}
		battles = append(battles, b)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return battles, nil
}


// Just ask the database whose turn it is
// - The database will return a secret key representing player1 or player2
func (m *BattleModel) CheckTurn(battleID, playerID int) string {
	var secretTurn string

	stmt := `SELECT secretTurn FROM Battles WHERE rowid = ? AND turn = ?`
	fmt.Println("secretTurn stmt:", stmt, "|", battleID, "|", playerID)
	_ = m.DB.QueryRow(stmt, battleID, playerID).Scan(&secretTurn)
	fmt.Println("secretTurn:", secretTurn)
	return secretTurn
}


// Just ask the database whose turn it is
// - The database will return a secret key representing player1 or player2
func (m *BattleModel) GetTurn(battleID int) (int, error) {
	var turn int = 0

	stmt := `SELECT turn FROM Battles rowid = ?`
	_ = m.DB.QueryRow(stmt, battleID).Scan(&turn)
	return turn, nil
}


// Update challenge - opponent has accepted
func (m *BattleModel) UpdateChallenge(player1 int, player2 int, player2Accepted bool, battleID int) (error) {
	stmt := `UPDATE Battles SET player2Accepted = ? WHERE rowid = ?`
	_, err := m.DB.Exec(stmt, player2Accepted, battleID)
	if err != nil {
		return err
	}
	return err
}


// Update turn - other player's turn
func (m *BattleModel) UpdateTurn(player1 int, player2 int, nextTurn int, battleID int, secretTurn []byte) error {
	// Swap the value of nextTurn
	if nextTurn == player1 {
		nextTurn = player2
	} else {
		nextTurn = player1
	}

	stmt := `UPDATE Battles SET secretTurn = ?, turn = ? WHERE rowid = ?`
	_, err := m.DB.Exec(stmt, secretTurn, nextTurn, battleID)
	if err != nil {
		return err
	}
	return nil
}
