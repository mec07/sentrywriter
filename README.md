# sentrywriter
Package sentrywriter is a wrapper around the sentry-go package and implements
the io.Writer interface. This allows us to send json formatted logs from
zerolog to Sentry (although there is no dependency on zerolog). You must
specify the log levels you want to send to Sentry.


## Example Usage
Here is a typical example, using zerolog. It is important to defer the
`sentryWriter.Flush` function because the messages are sent to Sentry
asynchronously.
```
package main

import (
	"github.com/mec07/sentrywriter"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	errLevel := sentrywriter.LogLevel{"error", sentry.LevelError}
	sentryWriter, err := sentrywriter.New(errLevel).WithUserID("userID").SetDSN("your-project-sentry-dsn")
	if err != nil {
		log.Error().Err(err).Str("dsn", "your-project-sentry-dsn").Msg("sentrywriter.SentryWriter.SetDSN")
		return
	}
	defer sentryWriter.Flush(2 * time.Second)

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	log.Logger = log.Output(zerolog.MultiLevelWriter(consoleWriter, sentryWriter))
}
```
