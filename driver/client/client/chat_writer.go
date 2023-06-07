package client

import (
	"bufio"
	"chat"
	"fmt"
	"os"
	"time"
)

func (c *client) WriteChat() error {
    fmt.Print("Type anything and press Enter: ")

    scanner := bufio.NewReader(os.Stdin)
    message, err := scanner.ReadString('\n')
    if err != nil {
        return err
    }
    message = message[:len(message)-1]

    chatMsg := &chat.ChatMessage{
        Username: c.username,
        Message: message,
        SendTime: time.Now(),
    }

    return c.chat.WriteToChat(chatMsg, c.chat.GetSessionChatName(c.curChat))
}
