package chat

import (
	"github.com/pkg/errors"
)

func (cs *ChatServer) isValidateChat(chat string) error {
	if chat != cs.GetSessionChatName(DAY_CHAT) &&
	   chat != cs.GetSessionChatName(NIGHT_CHAT) {
		return errors.Wrap(errNonExistingChat, chat)
	}
	
	return nil
}
