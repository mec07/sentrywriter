# Change Log

All notable changes to this project will be documented in this file. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.2] - 2020-10-24

### Added

- Added an example file: example/main.go

## [0.3.1] - 2020-10-20

### Fixed

- Protect the breadcrumbsLimit with a mutex

## [0.3.0] - 2020-10-19

### Changed

- Add SetClientOptions method to SentryWriter, i.e. just give the user the ability to toggle all the sentry.ClientOptions fields

## [0.2.0] - 2020-10-19

### Added

- Add the options to add breadcrumbs and attach stacktraces.

## [0.1.0] - 2020-10-19

### Changed

- Simplified log filtering. It is off by default but turns on as soon as any LogLevels are supplied.

## [0.0.3] - 2020-10-17

### Added

- Add some documentation to the Readme.

## [0.0.2] - 2020-10-17

### Fixed

- Fixed a bug with adding the User ID field.

## [0.0.1] - 2020-10-16

### Added

- Added the SentryWriter which implements the io.Writer interface.