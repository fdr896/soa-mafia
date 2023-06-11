package common

import "fmt"

type RabbitmqConnectionParams struct {
	User string
	Password string
	Hostname string
	Port string
}

func NewRabbitmqConnectionParams(user, password, hostname, port string) *RabbitmqConnectionParams {
	return &RabbitmqConnectionParams{
		User: user,
		Password: password,
		Hostname: hostname,
		Port: port,
	}
}

func GetRabbitmqConnectionUrl(connParams *RabbitmqConnectionParams) string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		connParams.User,
		connParams.Password,
		connParams.Hostname,
		connParams.Port)
}
