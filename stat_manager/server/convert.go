package server

import (
	"stat_manager/storage/database"
	"stat_manager/storage/filesystem"
)

//////////// FROM ////////////
func fromDbPlayer(p *database.Player) *PlayerInfo {
	return &PlayerInfo{
		Username: p.Username,
		Email: p.Email,
		Gender: database.GenderToString(p.Gender),
	}
}

func fromDbPlayers(ps []*database.Player) *PlayerInfos {
	var infos PlayerInfos
	infos.Players = make([]*PlayerInfo, 0)

	for _, player := range ps {
		infos.Players = append(infos.Players, fromDbPlayer(player))
	}

	return &infos
}

//////////// TO ////////////

func toDbPlayer(p *Player) *database.Player {
	return &database.Player{
		Username: p.Username,
		Email: p.Email,
		Gender: database.GenderFromString(p.Gender),
		AvatarFilename: filesystem.GetDefaultAvatar(),
	}
}

func toDbPlayerWithAvatar(p *Player, avatarFilename string) *database.Player {
	return &database.Player{
		Username: p.Username,
		Email: p.Email,
		Gender: database.GenderFromString(p.Gender),
		AvatarFilename: avatarFilename,
	}
}

func toDbPlayerForUpdate(username string, p *PlayerForUpdate) *database.Player {
	dp := database.Player{
		Username: username,
		Email: "",
		Gender: database.UNDEFINED,
	}

	if p.Email != nil {
		dp.Email = *p.Email
	}
	if p.Gender != nil {
		dp.Gender = database.GenderFromString(*p.Gender)
	}

	return &dp
}

func toDbPlayerForUpdateWithAvatar(username string, p *PlayerForUpdate, avatarFilename string) *database.Player {
	dp := toDbPlayerForUpdate(username, p)
	dp.AvatarFilename = avatarFilename

	return dp
}

func toDbPlayerStat(p *PlayerStat) *database.Player {
	return &database.Player{
		Username: p.Username,
		Email: "",
		Gender: database.UNDEFINED,
		SessionPlayed: p.SessionPlayed,
		GameWins: p.GameWins,
		GameLosts: p.GameLosts,
		TimePlayedMs: p.TimePlayedMs,
	}
}

