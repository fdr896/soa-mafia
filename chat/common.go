package chat

import (
	"fmt"

	"github.com/pkg/errors"
)

func getRabbitmqConnectionUrl(connParams *RabbitmqConnectionParams) string {
	return fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		connParams.user,
		connParams.password,
		connParams.hostname,
		connParams.port)
}

func (cs *ChatServer) isValidateChat(chat string) error {
	if chat != cs.GetSessionChatName(DAY_CHAT) &&
	   chat != cs.GetSessionChatName(NIGHT_CHAT) {
		return errors.Wrap(errNonExistingChat, chat)
	}
	
	return nil
}
