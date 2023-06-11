package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"

	_ "github.com/mattn/go-sqlite3"
	zlog "github.com/rs/zerolog/log"
)

type Players struct {
	db *sql.DB

	mutex sync.Mutex
}

func CreateOrReadPlayersDB(filename string) (*Players, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(createPlayersTableQuery); err != nil {
		return nil, err
	}

	return &Players{
		db: db,
	}, nil
}

func (ps *Players) CreatePlayer(ctx context.Context, p *Player) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	qResult, err := ps.db.ExecContext(
		ctx, createPlayerQuery,
		p.Username, p.Email, p.Gender, p.AvatarFilename,
		0, 0, 0, 0,
	)
	if err != nil {
		return err
	}

	var newId int64
	if newId, err = qResult.LastInsertId(); err != nil {
		return err
	}

	zlog.Info().Int64("id", newId).Msg("new player created")

	return nil
}

func (ps *Players) GetPlayerByUsername(ctx context.Context, username string) (*Player, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	row := ps.db.QueryRowContext(ctx, selectPlayerByUsernameQuery, username)

	var p Player
	err := row.Scan(&p.DbId,
		&p.Username, &p.Email, &p.Gender, &p.AvatarFilename,
		&p.SessionPlayed, &p.GameWins, &p.GameLosts, &p.TimePlayedMs)
	switch err {
	case nil:
		return &p, nil
	case sql.ErrNoRows:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (ps *Players) GetPlayersByUsernames(ctx context.Context, usernames []string) ([]*Player, error) {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	query := selectPlayersByUsernamesQuery(usernames)
	zlog.Debug().Str("query", query).Msg("select many")

	rows, err := ps.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	players := make([]*Player, 0)
	for rows.Next() {
		var p Player
		err := rows.Scan(&p.DbId,
			&p.Username, &p.Email, &p.Gender, &p.AvatarFilename,
			&p.SessionPlayed, &p.GameWins, &p.GameLosts, &p.TimePlayedMs)

		switch err {
		case nil:
			players = append(players, &p)
		default:
			return nil, err
		}
	}

	return players, nil
}

func (ps *Players) DeletePlayerByUsername(ctx context.Context, username string) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	_, err := ps.db.ExecContext(ctx, deletePlayerQuery, username)
	return err
}

func (ps *Players) UpdatePlayer(ctx context.Context, p *Player) error {
	ps.mutex.Lock()
	defer ps.mutex.Unlock()

	zlog.Debug().Interface("player", p).Msg("updating")

	fields := make([]string, 0)
	values := make([]string, 0)

	if len(p.Email) != 0 {
		fields = append(fields, "email")
		values = append(values, fmt.Sprintf("'%s'", p.Email))
	}
	if p.Gender != UNDEFINED {
		fields = append(fields, "gender")
		values = append(values, strconv.Itoa(p.Gender))
	}
	if len(p.AvatarFilename) != 0 {
		fields = append(fields, "avatar_filename")
		values = append(values, fmt.Sprintf("'%s'", p.AvatarFilename))
	}

	if p.SessionPlayed != 0 {
		fields = append(fields, "sessions_played")
		values = append(values, strconv.Itoa(p.SessionPlayed))
	}
	if p.GameWins != 0 {
		fields = append(fields, "game_wins")
		values = append(values, strconv.Itoa(p.GameWins))
	}
	if p.GameLosts != 0 {
		fields = append(fields, "game_losts")
		values = append(values, strconv.Itoa(p.GameLosts))
	}
	if p.TimePlayedMs != 0 {
		fields = append(fields, "time_playing_ms")
		values = append(values, strconv.Itoa(p.TimePlayedMs))
	}

	zlog.Debug().Interface("fields", fields).Msg("updating")
	zlog.Debug().Interface("values", values).Msg("updating")

	if len(fields) == 0 {
		return nil
	}

	query := updatePlayerInfoByUsernameQuery(p.Username, fields, values)
	zlog.Debug().Str("query", query).Msg("update")

	_, err := ps.db.ExecContext(ctx, query)
	return err
}
