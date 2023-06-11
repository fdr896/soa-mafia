package cli

import (
	"fmt"
	"os"
)


type interactor struct {
    commands chan *Command
}

func NewCliInteractor() *interactor {
    return &interactor{
        commands: make(chan *Command),
    }
}

func (i *interactor) Commands() <-chan *Command {
    return i.commands
}

func (i *interactor) Start() string {
    fmt.Printf(`
Hello!
You are connected to the grpc-mafia server! Here you can play mafia!!!
- Type 'start' to connect to random game session
- Type 'exit' to end the game
`)

    var cmd string
    for {
        _, err := fmt.Scan(&cmd)
        if err != nil {
            panic(err)
        }

        switch cmd {
        case "start":
            fmt.Print("Enter username (small latin letters, digits and underscores): ")
            return i.readUsername()
        case "exit":
            os.Exit(0)
        default:
            fmt.Println("Type 'start' or 'exit" )
        }
    }
}

func (i *interactor) readUsername() string {
    var username string
    for {
        _, err := fmt.Scan(&username)
        if err != nil {
            panic(err)
        }

        if isGoodUsername(username) {
            return username
        }
        fmt.Print("Invalid username, try again: ")
    }
}

func isGoodUsername(username string) bool {
    if len(username) == 0 {
        return false
    }

    for _, r := range username {
        if !('a' <= r && r <= 'z') &&
           !('0' <= r && r <= '9') &&
           !(r == '_') {
            return false
        }
    }

    return true
}

func (i *interactor) Print(msg string) {
    fmt.Println(msg)
}

func (i *interactor) PrintRules() {
    fmt.Println(`
[[Commands]]:
Voting:
    - 'vote' (enter voting mode to vote for mafia)
Chat:
    - 'read' (start reading current day chat)
    - 'read_all' (read all chat history)
    - 'read_last_n' (read last 'n' messages)
    - 'write' (write to game session chat)
Game Information:
    - 'role' (your role)
    - 'state' (game state)
    - 'nicks' (alive players' nicknames)
Interfaction With The Interface:
    - 'exit' (interrupt the game)
    - 'rules' (rules again)`)
}

func (i *interactor) StartPlaying(waitActionResp chan interface{}, nicknames *[]string) {
    var command string
    for {
        fmt.Print("\nCommand: ")
        _, err := fmt.Scanln(&command)
        if err != nil {
            continue
        }

        cmd, errorString := Parse(command, nicknames)
        if cmd == nil{
            fmt.Printf("Wrong command: %s\n", errorString)
        } else {
            i.commands <- cmd
            <-waitActionResp
        }
    }
}
