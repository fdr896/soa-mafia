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

func (i *interactor) Start(username string) {
    fmt.Printf(`
Hello, %s!
Your are connected to the grpc-mafia server! Here you can play mafia!!!
- Type 'start' to connect to random game session
- Type 'exit' to end the game
`, username)

    var cmd string
    for {
        _, err := fmt.Scan(&cmd)
        if err != nil {
            panic(err)
        }

        switch cmd {
        case "start":
            return
        case "exit":
            os.Exit(0)
        default:
            fmt.Println("Type 'start' or 'exit" )
        }
    }
}

func (i *interactor) Print(msg string) {
    fmt.Println(msg)
}

func (i *interactor) PrintRules() {
    fmt.Println(`
[[Commands]]:
Voting:
    - 'vote' (enter voting mode to vote for mafia)
Game Information:
    - 'role' (your role)
    - 'state' (game state)
    - 'nicks' (alive players' nicknames)
Interfaction With The Interface:
    - 'exit' (interrupt the game)
    - 'rules' (rules again)`)
}

func (i *interactor) StartPlaying(waitActionResp chan interface{}) {
    var command string
    for {
        fmt.Print("\nCommand: ")
        _, err := fmt.Scanln(&command)
        if err != nil {
            continue
        }

        cmd, errorString := Parse(command)
        if cmd == nil{
            fmt.Printf("Wrong command: %s\n", errorString)
        } else {
            i.commands <- cmd
            <-waitActionResp
        }
    }
}