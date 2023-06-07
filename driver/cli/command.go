package cli

import (
	"fmt"
	"strconv"
	"strings"
)

const (
    CMD_ROLE int = iota
    CMD_STATE
    CMD_NICKS
    CMD_VOTE
    CMD_EXIT
    CMD_RULES
    CMD_READ
    CMD_READ_ALL
    CMD_READ_LAST_N
    CMD_WRITE
)

var cmdNameToValue = map[string]int{
    "role": CMD_ROLE,
    "state": CMD_STATE,
    "nicks": CMD_NICKS,
    "vote": CMD_VOTE,
    "exit": CMD_EXIT,
    "rules": CMD_RULES,
    "read": CMD_READ,
    "read_all": CMD_READ_ALL,
    "read_last_n": CMD_READ_LAST_N,
    "write": CMD_WRITE,
}

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

func Parse(command string, nicknames *[]string) (*Command, string) {
    words := strings.Split(strings.TrimSpace(command), " ")

    if len(words) == 0 {
        return nil, "no command typed"
    }

    cmd := words[0]
    if cmdType, ok := cmdNameToValue[cmd]; ok {
        switch cmdType {
        case CMD_VOTE:
            return handleVoteCmd(nicknames)
        case CMD_READ_LAST_N:
            return handleReadLastN()
        default:
            return &Command{
                cmdType: cmdType,
                cmdArgs: words[1:],
            }, ""
        }

    } else {
        return nil, fmt.Sprintf("not available command: %s", cmd)
    }
}

func handleVoteCmd(nicknames *[]string) (*Command, string) {
    var suspect string
    for {
        fmt.Printf("Enter suspect username (%v): ", *nicknames)
        _, err := fmt.Scanln(&suspect)
        suspect = strings.TrimSpace(suspect)
        if len(suspect) > 0 && err == nil {
            break
        }
    }

    return &Command{
        cmdType: CMD_VOTE,
        cmdArgs: []string{suspect},
    }, ""
}

func handleReadLastN() (*Command, string) {
    var n string
    for {
        fmt.Print("How many last messages you want to see: ")
        _, err := fmt.Scanln(&n)
        n = strings.TrimSpace(n)
        if len(n) > 0 && err == nil {
            num, err := strconv.Atoi(n)
            if err == nil && num > 0 {
                break
            }
        }
    }

    return &Command{
        cmdType: CMD_READ_LAST_N,
        cmdArgs: []string{n},
    }, ""
}
