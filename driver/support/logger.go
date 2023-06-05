package support

import (
	"log"
	"os"
	"path"
	"path/filepath"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

const LOGS_DIR = ".logs"

func initStdoutLogger() {
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func InitServerLogger() {
    initStdoutLogger()
}

func InitClientLogger(username string, auto bool) error {
    if auto {
        initStdoutLogger()
        return nil
    }

    if err := os.MkdirAll(LOGS_DIR, 0744); err != nil {
        panic(err)
    }

    logFilename := path.Join(LOGS_DIR, username + ".log")
    logFile, err := os.OpenFile(logFilename, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
    if err != nil {
        return err
    }

    absLogsPath, err := filepath.Abs(logFilename)
    if err != nil {
        return err
    }
    log.Printf("Writing logs to file: %s\n", absLogsPath)

    zlog.Logger = zerolog.New(logFile).With().Timestamp().Logger()

    return nil
}
