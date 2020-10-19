# sentrywriter
Package sentrywriter is a wrapper around the sentry-go package and implements
the io.Writer interface. This allows us to send logs from zerolog (or some
other logging package that accepts the io.Writer interface) and send them to
Sentry (there is no dependency on zerolog in this package).

There is a mechanism in this package to filter json formatted logs (we
normally only want to send errors to Sentry, rather than all logs). When you
supply a `LogLevel` to the writer, you tell it to turn on filtering and to
check that all json formatted logs have a `"level"` field (you can change
this default using the `WithLevelFieldName` function) and that it matches one
of the supplied `LogLevel`s.

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
