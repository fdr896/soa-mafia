package database

const (
	MALE = 1
	FEMALE = 2
	UNDEFINED = -1
)

type Player struct {
	DbId int `json:"db_id"`

	// personal data
	Username       string `json:"username"`
	Email          string `json:"email"`
	Gender         int    `json:"gender"` // [MALE|FEMALE]
	AvatarFilename string `json:"avatar_filename"`

	// statistics
	SessionPlayed int `json:"session_played"`
	GameWins      int `json:"game_wins"`
	GameLosts     int `json:"game_losts"`
	TimePlayedMs  int `json:"time_played_ms"`
}

func GenderToString(gender int) string {
	switch gender {
	case MALE:
		return "male"
	case FEMALE:
		return "female"
	default:
		return "undefined"
	}
}

func GenderFromString(gender string) int {
	switch gender{
	case "male":
		return MALE
	case "female":
		return FEMALE
	default:
		return UNDEFINED
	}
}
