package client

import (
	"chat"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/eiannone/keyboard"
	zlog "github.com/rs/zerolog/log"
)

func (c *client) StartReadingChat(errChan chan<- error) {
	c.msgsAccess = sync.Mutex{}

	msgsChan, err := c.chat.StartReadFromChat()
	if err != nil {
		zlog.Error().Err(err).Msg("failed to start consuming from chat")
		errChan <- err
		return
	}
	zlog.Info().Str("username", c.username).Msg("client start reading from chat")

	for msg := range msgsChan {
		var chatMsg chat.ChatMessage
		if err := json.Unmarshal(msg.Body, &chatMsg); err != nil {
			zlog.Error().Err(err).Msg("failed to parse chat message")
			errChan <- err
			return
		}
		zlog.Debug().Str("username", c.username).Str("chat", c.curChat).Interface("msg", chatMsg).Msg("new chat msg")

		c.msgsAccess.Lock()
		c.chatsMsgs[c.curChat] = append(c.chatsMsgs[c.curChat], &chatMsg)
		c.msgsAccess.Unlock()

		go func() {
			c.newMsgs <- struct{}{}
		} ()
	}
}

type MsgFilter func (*chat.ChatMessage) bool

func (c *client) ReadChatSession() error {
	lastSessionTime := c.lastSessionTime

	return c.readChat(func(msg *chat.ChatMessage) bool {
		return msg.SendTime.After(lastSessionTime)
	})
}

func (c *client) ReadChatAll() error {
	var printed int

	allMsgs := c.readUnprinted(&printed)

	for _, msg := range allMsgs {
		prettyPrint(msg)
	}

	return nil
}

func (c *client) ReadChatLastN(n int) error {
	var printed int

	allMsgs := c.readUnprinted(&printed)

	leftBound := func() int {
		if len(allMsgs) >= n {
			return len(allMsgs) - n
		} else {
			return 0
		}
	}()
	zlog.Info().Int("from", leftBound).Int("len", len(allMsgs)).Int("n", n).Msg("reading")

	for _, msg := range allMsgs[leftBound:] {
		prettyPrint(msg)
	}

	return nil
}

func (c *client) readChat(filter MsgFilter) error {
    stopReading := make(chan interface{}, 1)

    go func() {
		var printed int

		prevMsgs := c.readUnprinted(&printed)

		for _, msg := range prevMsgs {
			if filter(msg) {
				prettyPrint(msg)
			}
		}

		for {
			select {
			case <-stopReading:
				return
			case <-c.newMsgs:
				newMsgs := c.readUnprinted(&printed)
				for _, msg := range newMsgs {
					if filter(msg) {
						prettyPrint(msg)
					}
				}
			}
		}
    }()

    fmt.Println("Press any key to stop reading...")
    keyboard.GetSingleKey()

    stopReading <- struct{}{}

    return nil
}

func (c *client) readUnprinted(printed *int) []*chat.ChatMessage {
	zlog.Info().Str("username", c.username).Msg("reading unprinted")
	var unprintedMsgs []*chat.ChatMessage

	var startPos int
	if *printed == -1 {
		startPos = 0
	} else {
		startPos = *printed
	}

	c.msgsAccess.Lock()
	zlog.Debug().Str("username", c.username).Interface("msgs", c.chatsMsgs[c.curChat]).Int("start pos", startPos).Msg("chat msgs")
	unprintedMsgs = make([]*chat.ChatMessage, len(c.chatsMsgs[c.curChat][startPos:]))
	copy(unprintedMsgs, c.chatsMsgs[c.curChat][startPos:])
	zlog.Debug().Str("username", c.username).Interface("msgs", unprintedMsgs).Msg("copied")
	c.msgsAccess.Unlock()

	if *printed == -1 {
		*printed = len(unprintedMsgs)
	} else {
		*printed += len(unprintedMsgs)
	}

	return unprintedMsgs
}

func prettyPrint(msg *chat.ChatMessage) {
	fmt.Printf(
		"[%s] %s: %s\n",
		msg.SendTime.Format(time.RFC3339),
		msg.Username,
		msg.Message,
	)
}
