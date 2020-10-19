# sentrywriter
Package sentrywriter is a wrapper around the sentry-go package and implements
the io.Writer interface. This allows us to send logs from zerolog (or some
other logging package that accepts the io.Writer interface) and send them to
Sentry (there is no dependency on zerolog in this package).

There is a mechanism in this package to filter json formatted logs (we
normally only want to send errors to Sentry, rather than all logs). For
example, let's say you supply the writer with a `LogLevel`:
```
errorLevel := sentrywriter.LogLevel{
	MatchingString:"error",
	SentryLevel: sentry.ErrorLevel,
}
writer := sentrywriter.New(errorLevel)
```
The `writer` now has filtering turned on and when it next receives a log, it
json decodes it and checks the `"level"` field (you can change this default
using the `WithLevelFieldName` method) matches `"error"`. If it matches then
it sets the sentry level to `sentry.ErrorLevel` and sends the message to
Sentry. If no `LogLevel`s are provided then filtering is not turned on.
Multiple `LogLevel`s can be supplied both at instantiation time and at a
later point, for example:
```
errorLevel := sentrywriter.LogLevel{
	MatchingString: "error",
	SentryLevel: sentry.ErrorLevel,
}
fatalLevel := sentrywriter.LogLevel{
	MatchingString: "fatal",
	SentryLevel: sentry.FatalLevel,
}
writer := sentrywriter.New(errorLevel, fatalLevel)

warningLevel := sentrywriter.LogLevel{
	MatchingString: "warning",
	SentryLevel: sentry.WarningLevel,
}
writer.WithLogLevel(warningLevel)
```

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
	errorLevel := sentrywriter.LogLevel{"error", sentry.LevelError}
	sentryWriter, err := sentrywriter.New(errorLevel).WithUserID("userID").SetDSN("your-project-sentry-dsn")
	if err != nil {
		log.Error().Err(err).Str("dsn", "your-project-sentry-dsn").Msg("sentrywriter.SentryWriter.SetDSN")
		return
	}
	defer sentryWriter.Flush(2 * time.Second)

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	log.Logger = log.Output(zerolog.MultiLevelWriter(consoleWriter, sentryWriter))
}
```
