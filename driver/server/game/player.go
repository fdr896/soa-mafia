package game

const (
	MAFIA = 0
	COMISSAR = 1
	CIVILIAN = 2
	SPIRIT = 3
	UNDEFINED = 4
)

type player struct {
	id string
	nickname string
	role int // [MAFIA|COMISSAR|CIVILLIAN|SPIRIT]
}

func (p *player) getRole() string {
    switch p.role {
    case MAFIA:
        return "mafia"
    case COMISSAR:
        return "comissar"
    case CIVILIAN:
        return "civilian"
    case SPIRIT:
        return "spirit"
    case UNDEFINED:
        return "undefined"
    default:
        panic("unknown role")
    }
}

func (p *player) isAlive() bool {
    return !p.isSpirit()
}

func (p *player) isSpirit() bool {
	return p.role == SPIRIT
}
