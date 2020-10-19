/*
Package sentrywriter is a wrapper around the sentry-go package and implements
the io.Writer interface. This allows us to send logs from zerolog (or some
other logging package that accepts the io.Writer interface) and send them to
Sentry (there is no dependency on zerolog in this package).

There is a mechanism in this package to filter json formatted logs (we
normally only want to send errors to Sentry, rather than all logs). For
example, let's say you supply the writer with a `LogLevel`:
    errorLevel := sentrywriter.LogLevel{
    	MatchingString:"error",
    	SentryLevel: sentry.ErrorLevel,
    }
    writer := sentrywriter.New(errorLevel)

The `writer` now has filtering turned on and when it next receives a log, it
json decodes it and checks the `"level"` field (you can change this default
using the `WithLevelFieldName` method) matches `"error"`. If it matches then
it sets the sentry level to `sentry.ErrorLevel` and sends the message to
Sentry. Multiple `LogLevel`s can be supplied both at instantiation time and
at a later point, for example:
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

If no `LogLevel`s are provided then filtering is not turned on.

Here is a typical example, using zerolog. It is important to defer the
`sentryWriter.Flush` function because the messages are sent to Sentry
asynchronously.

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
	    defer sentryWriter.Flush()

	    consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	    log.Logger = log.Output(zerolog.MultiLevelWriter(consoleWriter, sentryWriter))
    }
*/
package sentrywriter

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
)

// SentryClient is an interface which represents the sentry-go package client.
type SentryClient interface {
	CaptureMessage(message string, hint *sentry.EventHint, scope sentry.EventModifier) *sentry.EventID
	Flush(timeout time.Duration) bool
}

// LogLevel is used to match the log level that you're using and then map it
// into a Sentry log level. For example, you may be logging at level "error",
// which corresponds to sentry.LevelError, so that would correspond to:
//     levelError := LogLevel{"error", sentry.LevelError}
//
// See https://godoc.org/github.com/getsentry/sentry-go#Level for the possible
// Sentry log levels.
type LogLevel struct {
	MatchingString string
	SentryLevel    sentry.Level
}

// SentryWriter implements the io.Writer interface. It is a wrapper over the
// sentry-go client and sends the supplied logs of the specified log level to
// Sentry. It assumes that the logs are json encoded. Writes are asynchronous,
// so remember to call Flush before exiting the program.
type SentryWriter struct {
	mu             sync.RWMutex
	client         SentryClient
	scope          *sentry.Scope
	logLevels      []LogLevel
	filterLogsFlag bool
	levelFieldName string
}

// New returns a pointer to the SentryWriter, with the specified log levels set.
// The SentryWriter will write logs which match any of the supplied logs to
// Sentry. The default field that is checked for the log level is "level". For
// example:
//     writer := sentrywriter.New(sentrywriter.LogLevel{"error", sentry.LevelError})
func New(logLevels ...LogLevel) *SentryWriter {

	// The sentry-go package
	writer := SentryWriter{
		levelFieldName: "level",
		scope:          sentry.NewScope(),
		logLevels:      logLevels,
	}
	if len(logLevels) > 0 {
		writer.turnOnFilterLogsFlag()
	}
	return &writer
}

// SetDSN sets the DSN for the Sentry client. For example:
//     writer, err := sentrywriter.New().SetDSN(dsn)
func (s *SentryWriter) SetDSN(DSN string) (*SentryWriter, error) {
	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn: DSN,
	})
	if err != nil {
		return nil, errors.Wrap(err, "sentry.NewClient")
	}

	s.client = client
	return s, nil
}

// WithLogLevel adds a LogLevel that triggers an event to be sent to Sentry. For
// example:
//     writer := sentrywriter.New().WithLogLevel(sentrywriter.LogLevel{"error", sentry.LevelError})
func (s *SentryWriter) WithLogLevel(logLevel LogLevel) *SentryWriter {
	s.addLogLevel(logLevel)

	if !s.shouldFilterLogs() {
		s.turnOnFilterLogsFlag()
	}

	return s
}

func (s *SentryWriter) addLogLevel(logLevel LogLevel) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logLevels = append(s.logLevels, logLevel)
}

// WithLevelFieldName allows you to change the log level field name from the
// default of "level" to whatever you are using. For example:
//     writer := sentrywriter.New().WithLevelFieldName("log_level")
func (s *SentryWriter) WithLevelFieldName(name string) *SentryWriter {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.levelFieldName = name
	return s
}

func (s *SentryWriter) getLevelFieldName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.levelFieldName
}

// WithUserID sets a user ID that will be reported alongside each Sentry event.
// This is helpful for code that runs on client machines. For example:
//     writer := sentrywriter.New().WithUserID("userID")
func (s *SentryWriter) WithUserID(userID string) *SentryWriter {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.scope.SetUser(sentry.User{ID: userID})
	return s
}

// WithClient allows you to substitute the client that is being used, rather
// than the default client from the sentry-go package. For example:
//     writer := sentrywriter.New().WithClient(client)
// where client implements the SentryClient interface.
func (s *SentryWriter) WithClient(client SentryClient) *SentryWriter {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.client = client
	return s
}

// Write is the implementation of the io.Writer interface. It checks if the log
// is at one of the preset log levels and if so it writes it to Sentry.
func (s *SentryWriter) Write(log []byte) (int, error) {
	if s.client == nil {
		return 0, errors.New("no Sentry client supplied")
	}

	scope := s.getScope()

	if s.shouldFilterLogs() {
		var eventMap map[string]json.RawMessage
		if err := json.Unmarshal(log, &eventMap); err != nil {
			return 0, errors.Wrap(err, "json.Unmarshal log")
		}
		var level string
		if err := json.Unmarshal(eventMap[s.getLevelFieldName()], &level); err != nil {
			return 0, errors.Wrapf(err, `json.Unmarshal eventMap["%s"]`, s.getLevelFieldName())
		}

		logLevel, found := s.findMatchingLogLevel(level)
		if !found {
			return len(log), nil
		}

		scope.SetLevel(logLevel.SentryLevel)
	}

	s.client.CaptureMessage(string(log), nil, scope)

	return len(log), nil
}

func (s *SentryWriter) getScope() *sentry.Scope {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.scope.Clone()
}

func (s *SentryWriter) findMatchingLogLevel(level string) (LogLevel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, logLevel := range s.logLevels {
		if logLevel.MatchingString == level {
			return logLevel, true
		}
	}
	return LogLevel{}, false
}

// Flush initiates the Flush method of the underlying Sentry client. Call this
// before exiting your program. The provided timeout is the maximum length of
// time to block until all the logs have been sent to Sentry. It returns false
// if the timeout is reached, which may signify that not all messages were sent
// to Sentry.
func (s *SentryWriter) Flush(timeout time.Duration) bool {
	return s.client.Flush(timeout)
}

func (s *SentryWriter) shouldFilterLogs() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.filterLogsFlag
}

func (s *SentryWriter) turnOnFilterLogsFlag() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.filterLogsFlag = true
}
