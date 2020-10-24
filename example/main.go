package main

import (
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/mec07/sentrywriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	errorLevel := sentrywriter.LogLevel{
		MatchingString: "error",
		SentryLevel:    sentry.LevelError,
	}
	options := sentry.ClientOptions{
		Dsn:              "https://your-dsn-replaces-this@idstuff.ingest.sentry.io/1234567",
		AttachStacktrace: true,
		Environment:      "your-environment",
		Release:          "the-version-of-this-release",
	}
	sentryWriter, err := sentrywriter.New(errorLevel).WithUserID("userID").SetClientOptions(options)
	if err != nil {
		log.Error().Err(err).Str("dsn", "your-project-sentry-dsn").Msg("sentrywriter.SentryWriter.SetDSN")
		return
	}
	defer sentryWriter.Flush(2 * time.Second)

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	log.Logger = log.Output(zerolog.MultiLevelWriter(consoleWriter, sentryWriter))
}
