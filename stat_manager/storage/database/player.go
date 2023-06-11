package database

const (
	MALE = 1
	FEMALE = 2
	UNDEFINED = -1
)

type Player struct {
	dbId int

	// personal data
	Username string
	Email string
	Gender int // [MALE|FEMALE]
	AvatarFilename string

	// statistics
	SessionPlayed int
	GameWins int
	GameLosts int
	TimePlayedMs int
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
