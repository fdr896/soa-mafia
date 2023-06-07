package game

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/pkg/errors"
	zlog "github.com/rs/zerolog/log"
)

const (
	DAY = 0
	NIGHT = 1
)

func NewGameSession(id string, gamePlayers, mafias int) *GameSession {
	return &GameSession{
		Id: id,
        GamePlayers: gamePlayers,
        Mafias: mafias,

        mafiaNicknames: make([]string, 0),

		players: make(map[string]*player),
		playerByNickname: make(map[string]*player),
	}
}

type GameSession struct {
	Id string
    GamePlayers int
    Mafias int

	currentDay int
	timeOfDay int // [DAY|NIGHT]

	alivePlayers int
	votes int // resets after each day
	votesAgainstPlayer map[string]int
    mafiaNicknames []string

	players map[string]*player
	playerByNickname map[string]*player

    // night state
    mafiaVote string
    aliveMafiaVotes int
    commissarVote string
    mafiaContradicts bool
    isComissarAlive bool
    commisarFoundMafia bool
    commisarDesideIfPublish bool
}

// Needs to find a not started game session
func (gs *GameSession) IsStarted() bool {
	return len(gs.players) == gs.GamePlayers
}

func (gs *GameSession) IsMafia(id string) bool {
    return gs.players[id].role == MAFIA
}

// Returns true if session is fullfilled
// (implied to start the game after true was returned)
func (gs *GameSession) AddPlayer(id string, nickname string) (bool, bool) {
    if _, ok := gs.playerByNickname[nickname]; ok {
        return false, true
    }

	player := &player{
		id: id,
		nickname: nickname,
		role: UNDEFINED,
	}
	gs.players[id] = player
	gs.playerByNickname[nickname] = player
	gs.alivePlayers += 1

	return len(gs.players) == gs.GamePlayers, false
}

func (gs *GameSession) GetTimeOfDay() int {
    return gs.timeOfDay
}

func (gs *GameSession) GetDay() int {
    return gs.currentDay
}

func (gs *GameSession) GetPlayerRole(id string) int {
	return gs.players[id].role
}

func (gs *GameSession) GetPlayerNickname(id string) string {
    return gs.players[id].nickname
}

func (gs *GameSession) GetPlayerRoleString(id string) string {
	return gs.players[id].getRole()
}

func (gs *GameSession) GetAlivePlayersCount() int {
    alive := 0
    for _, player := range gs.players {
        if player.isAlive() {
            alive += 1
        }
    }

    return alive
}

// Returnes only alive players' nicknames
func (gs *GameSession) GetPlayerNicknames() []string {
	nicknames := make([]string, 0)
	for _, player := range gs.players {
		if player.isAlive() {
			nicknames = append(nicknames, player.nickname)
		}
	}
    sort.Slice(nicknames, func(i, j int) bool {
        return nicknames[i] < nicknames[j]
    })

	return nicknames
}

// Returnes all players' nicknames
func (gs *GameSession) GetAllPlayerNicknames() []string {
	nicknames := make([]string, 0)
	for _, player := range gs.players {
		nicknames = append(nicknames, player.nickname)
	}

	return nicknames
}

func (gs *GameSession) GetMafiaNicknames() []string {
    return gs.mafiaNicknames
}

func (gs *GameSession) GetPlayerIdByNickname(nickname string) string {
    return gs.playerByNickname[nickname].id
}

func (gs *GameSession) StartGame() {
	// initialize fields
	gs.currentDay = 1
	gs.timeOfDay = DAY
	gs.votes = 0
	gs.votesAgainstPlayer = make(map[string]int)
    gs.mafiaVote = ""
    gs.commissarVote = ""
    gs.aliveMafiaVotes = 0
    gs.mafiaContradicts = false
    gs.isComissarAlive = true
    gs.commisarFoundMafia = false
    gs.commisarDesideIfPublish = false

	// randomly distribute roles
	ids := make([]string, 0)
	for id := range gs.players {
		ids = append(ids, id)
	}

	rand.Shuffle(len(ids), func(i, j int) {
		ids[i], ids[j] = ids[j], ids[i]
	})

    // first player is commissar
	// gs.mafias players is mafia
	// rest are civilians
    gs.players[ids[0]].role = COMISSAR
    for _, playerId := range ids[1:gs.Mafias + 1]{
	    gs.players[playerId].role = MAFIA
    }
	for _, player := range gs.players {
        if player.role != MAFIA && player.role != COMISSAR {
			player.role = CIVILIAN

        }
	}

    // print roles
    for id, player := range gs.players {
        zlog.Info().
        Str("id", id).
        Str("nick", player.nickname).
        Str("role", player.getRole()).
        Msg("player")
    }

    for _, player := range gs.players {
        if player.role == MAFIA {
            gs.mafiaNicknames = append(gs.mafiaNicknames, player.nickname)
        }
    }

}

// Returns game status
func (gs *GameSession) StartDay() int {
    status := gs.isGameEnded()
    if status != NOT_FINISHED {
        return status
    }

	gs.currentDay += 1
	gs.timeOfDay = DAY

    gs.votes = 0
    gs.votesAgainstPlayer = make(map[string]int)
    gs.mafiaVote = ""
    gs.commissarVote = ""
    gs.aliveMafiaVotes = 0
    gs.mafiaContradicts = false
    gs.commisarFoundMafia = false
    gs.commisarDesideIfPublish = false

    return NOT_FINISHED
}

func (gs *GameSession) StartNight() {
	gs.timeOfDay = NIGHT
}

const (
    MAFIAN_WON = 0
    CIVILIAN_WON = 1
    NOT_FINISHED = 2
)

func (gs *GameSession) isGameEnded() int {
    aliveMafias := gs.getAliveMafias()
    if aliveMafias == 0 {
        return CIVILIAN_WON
    }

    civiliansCnt := 0
    for _, player := range gs.players {
        if player.role == CIVILIAN {
            civiliansCnt += 1
        }
    }

    if civiliansCnt <= aliveMafias {
        return MAFIAN_WON
    }

    return NOT_FINISHED
}

func (gs *GameSession) getAliveMafias() int {
    aliveMafias := 0
    for _, nickname := range gs.mafiaNicknames {
        if gs.playerByNickname[nickname].isAlive() {
            aliveMafias += 1
        }
    }

    return aliveMafias
}

type MorningSummary struct {
    KilledPlayerNickname string
    KilledByMafiaPlayerNickname string
    CommissarInvestigationResult string
}

func (gs *GameSession) GetMorningSummary() *MorningSummary {
    if gs.currentDay == 1 {
        return &MorningSummary{
            KilledPlayerNickname: "nobody was killed",
            KilledByMafiaPlayerNickname: "nobody was killed",
            CommissarInvestigationResult: "mafia was not found",
        }
    }

    killedPlayerId, killed := gs.getMostFrequentVote()
    killedPlayerByMafiaId := gs.mafiaVote
    publishInvestigation := gs.commisarFoundMafia

    var killedPlayer string
    if killed {
        gs.killPlayer(killedPlayerId)
        killedPlayer = gs.GetPlayerNickname(killedPlayerId)
    } else {
        killedPlayer = killedPlayerId
    }

    var killedPlayerByMafia string
    if !gs.mafiaContradicts {
        gs.killPlayer(killedPlayerByMafiaId)
        killedPlayerByMafia = gs.GetPlayerNickname(killedPlayerByMafiaId)
    } else {
        killedPlayerByMafia = "nobody was killed (mafias did not agreed with each other)"
    }

    return &MorningSummary{
        KilledPlayerNickname: killedPlayer,
        KilledByMafiaPlayerNickname: killedPlayerByMafia,
        CommissarInvestigationResult: gs.commissarMessage(publishInvestigation),
    }
}

// Returns true if somebody was killed
func (gs *GameSession) getMostFrequentVote() (string, bool) {
    ids := make([]string, 0)
    for id := range gs.players {
        ids = append(ids, id)
    }

    sort.Slice(ids, func(i, j int) bool {
        return gs.votesAgainstPlayer[ids[i]] > gs.votesAgainstPlayer[ids[j]]
    })

    zlog.Info().Interface("ids", ids).Msg("player ids")
    zlog.Info().Interface("votes", gs.votesAgainstPlayer).Msg("votes")
    zlog.Info().Int("vote 1", gs.votesAgainstPlayer[ids[0]]).Msg("most frequent")
    zlog.Info().Int("vote 2", gs.votesAgainstPlayer[ids[1]]).Msg("second most frequent")

    if gs.votesAgainstPlayer[ids[0]] == gs.votesAgainstPlayer[ids[1]] {
        return "nobody was killed", false
    } else {
        return ids[0], true
    }
}

func (gs *GameSession) killPlayer(id string) {
    player := gs.players[id]
    if player.role == COMISSAR {
        gs.isComissarAlive = false
    }
    if player.role != SPIRIT {
        gs.alivePlayers -= 1
        player.role = SPIRIT
    }
}

func (gs *GameSession) commissarMessage(found bool) string {
    if found {
        return fmt.Sprintf("mafia was found: %s", gs.GetPlayerNickname(gs.commissarVote))
    } else {
        return "mafia was not found"
    }
}

type NightInfo struct {
    MafiaIds []string
    ComissarId string
    CivilianIds []string
}

func (gs *GameSession) GetNightInfo() *NightInfo {
    nightInfo := NightInfo{
        MafiaIds: make([]string, 0),
        CivilianIds: make([]string, 0),
    }

    for _, player := range gs.players {
        if player.isSpirit() {
            continue
        }

        switch player.role {
        case MAFIA:
            nightInfo.MafiaIds = append(nightInfo.MafiaIds, player.id)
        case COMISSAR:
            nightInfo.ComissarId = player.id
        case CIVILIAN:
            nightInfo.CivilianIds = append(nightInfo.CivilianIds, player.id)
        }
    }

    return &nightInfo
}


// Returns
// 1. (false, nil) if not all players voted
// 2. (true, nil) if all players voted and night started
//
// Vote call succedes iff
// 1. both players are not spirits
// 2. current time of day is day 
func (gs *GameSession) Vote(voterId string, suspectNickname string) (bool, error) {
	voter := gs.players[voterId]
	suspect, ok2 := gs.playerByNickname[suspectNickname]

    if !ok2 {
        return false, fmt.Errorf("no such player: %s", suspectNickname)
    }

	errPref := zlog.Error().Str("voter", voter.nickname).Str("suspect", suspect.nickname)

	if voter.isSpirit() {
		err := errors.Wrap(ErrNotPermitted, "spirit can not vote")
		errPref.Err(err).Msg("failed to vote")
		return false, err
	}
	if suspect.isSpirit() {
		err := errors.Wrap(ErrNotPermitted, "suspect is spirit")
		errPref.Err(err).Msg("failed to vote")
		return false, err
	}

	if gs.timeOfDay == NIGHT {
		err := errors.Wrap(ErrNotPermitted, "cannot vote in night")
		errPref.Err(err).Msg("failed to vote")
		return false, err
	}

    zlog.Info().Str("voter", voter.nickname).Str("suspect", suspect.nickname).Msg("vote")

	gs.votes += 1
    gs.votesAgainstPlayer[suspect.id] += 1
	if gs.votes == gs.alivePlayers {
		gs.StartNight()
		return true, nil
	}

	return false, nil
}

// Returns true if both mafia and commissar voted
func (gs *GameSession) AcceptMafiaVote(mafiaId, playerId string) (bool, error) {
	errPref := zlog.Error().Str("mafia", mafiaId).Str("player", playerId)

    player, ok := gs.players[playerId]
    if !ok {
        return false, fmt.Errorf("no player with id: %s", playerId)
    }

    if player.isSpirit() {
		err := errors.Wrap(ErrNotPermitted, "mafia cannot kill spirit")
		errPref.Err(err).Msg("failed to kill")
        return false, err
    }

    if len(gs.mafiaVote) > 0 && playerId != gs.mafiaVote {
        gs.mafiaContradicts = true
    }
    gs.mafiaVote = playerId
    gs.aliveMafiaVotes += 1

    if gs.aliveMafiaVotes < gs.getAliveMafias() {
        return false, nil
    }

    if gs.isComissarAlive && !gs.commisarDesideIfPublish {
        return false, nil
    }

    if !gs.isComissarAlive || (gs.isComissarAlive && len(gs.commissarVote) > 0) {
        return true, nil
    }
    return false, nil
}

func (gs *GameSession) AcceptCommissarVote(commissarId, playerId string) error {
	errPref := zlog.Error().Str("commissar", commissarId).Str("player", playerId)

    player, ok := gs.players[playerId]
    if !ok {
        return fmt.Errorf("no player with id: %s", playerId)
    }

    if player.isSpirit() {
		err := errors.Wrap(ErrNotPermitted, "commissar can not investigate spirit")
		errPref.Err(err).Msg("failed to investigate")
        return err
    }

    gs.commissarVote = playerId
    if player.role != MAFIA {
        gs.commisarDesideIfPublish = true
    }

    return nil
}

// If invoked, message with mafia name will be published in the next day
// Returns true if both commissar and mafia voted
func (gs *GameSession) CommissarFoundMafia(publisResult bool) bool {
    gs.commisarFoundMafia = publisResult
    gs.commisarDesideIfPublish = true

    return gs.aliveMafiaVotes == gs.getAliveMafias()
}
