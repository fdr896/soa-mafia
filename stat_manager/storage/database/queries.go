package database

import (
	"fmt"
	"strings"
)

const createPlayersTableQuery =
`
BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS players (
	id INTEGER PRIMARY KEY AUTOINCREMENT,

	username TEXT NOT NULL,
	email TEXT NOT NULL,
	gender INTEGER NOT NULL,
	avatar_filename TEXT NOT NULL,

	sessions_played INTEGER NOT NULL,
	game_wins INTEGER NOT NULL,
	game_losts INTEGER NOT NULL,
	time_playing_ms INTEGER NOT NULL
);
COMMIT;
`

const createPlayerQuery =
`
BEGIN TRANSACTION;
INSERT INTO players VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, ?);
COMMIT;
`

const deletePlayerQuery =
`
BEGIN TRANSACTION;
DELETE FROM players WHERE username = ?;
COMMIT;
`

const selectPlayerByUsernameQuery =
`
SELECT * FROM players WHERE username = ?
`

func selectPlayersByUsernamesQuery(usernames []string) string {
	for i := range usernames {
		usernames[i] = fmt.Sprintf("'%s'", usernames[i])
	}
	return fmt.Sprintf(
		"SELECT * FROM players WHERE username IN (%s)",
		strings.Join(usernames, ","))
}

func updatePlayerInfoByUsernameQuery(username string, fields, values []string) string {
	query := "UPDATE players SET "

	for i := range fields {
		if fields[i] == "sessions_played" ||
		   fields[i] == "game_wins" ||
		   fields[i] == "game_losts" ||
		   fields[i] == "time_playing_ms" {
			query += fmt.Sprintf("%s = %s + %s", fields[i], fields[i], values[i])
		} else {
			query += fmt.Sprintf("%s = %s", fields[i], values[i])
		}
		if i != len(fields) - 1 {
			query += ","
		} else {
			query += " "
		}
	}

	query += fmt.Sprintf("WHERE username = '%s'", username)
	
	return query
}
