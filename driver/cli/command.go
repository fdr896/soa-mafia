package cli

import (
	"fmt"
	"strings"
)

const (
    CMD_ROLE = 0
    CMD_STATE = 1
    CMD_NICKS = 2
    CMD_VOTE = 3
    CMD_EXIT = 4
    CMD_RULES = 5
)

type Command struct {
    cmdType int
    cmdArgs []string
}

func (c *Command) GetType() int {
    return c.cmdType
}

func (c *Command) GetArg(pos int) string {
    return c.cmdArgs[pos]
}

func Parse(command string) (*Command, string) {
    words := strings.Split(strings.TrimSpace(command), " ")

    if len(words) == 0 {
        return nil, "no command typed"
    }

    cmd := words[0]
    if cmdType, ok := cmdNameToValue[cmd]; ok {
        if cmdType == CMD_VOTE {
            var suspect string
            for {
                fmt.Print("Enter suspect username: ")
                _, err := fmt.Scanln(&suspect)
                suspect = strings.TrimSpace(suspect)
                if len(suspect) > 0 && err == nil {
                    break
                }
            }

            return &Command{
                cmdType: cmdType,
                cmdArgs: []string{suspect},
            }, ""
        }

        return &Command{
            cmdType: cmdType,
            cmdArgs: words[1:],
        }, ""
    } else {
        return nil, fmt.Sprintf("not available command: %s", cmd)
    }
}

var cmdNameToValue = map[string]int{
    "role": CMD_ROLE,
    "state": CMD_STATE,
    "nicks": CMD_NICKS,
    "vote": CMD_VOTE,
    "exit": CMD_EXIT,
    "rules": CMD_RULES,
}