package client

import (
	"fmt"
	"strings"
)

// Eventually returnes nickname of player
// that mafia wants to kill
func (c *client) StartMafiaChat() (string, error) {
    fmt.Println(`
You can communicate with other mafias and kill the player
[[Chat Commands]]:
    - 'read' (start reading chat)
    - 'write' (write to chat)
    - 'kill' (choose player that you want to kill)`)

	var cmd string
	for {
		fmt.Print("Command: ")
		_, err := fmt.Scanln(&cmd)
		cmd = strings.TrimSpace(cmd)
		if err != nil {
			return "", err
		}

		switch cmd {
		case "read":
			if err := c.ReadChatSession(); err != nil {
				return "", err
			}
		case "write":
			if err := c.WriteChat(); err != nil {
				return "", err
			}
		case "kill":
			var nickname string
			for {
				fmt.Print("Type nickname to kill: ")
				_, err := fmt.Scanln(&nickname)
				nickname = strings.TrimSpace(nickname)
				if err == nil && len(nickname) > 0 {
					break
				}
			}
			nickname = strings.Split(strings.TrimSpace(nickname), " ")[0]

			return nickname, nil
		default:
			continue
		}
	}
}
