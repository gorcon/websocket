# Changelog
All notable changes to this project will be documented in this file.

**ATTN**: This project uses [semantic versioning](http://semver.org/).

## [Unreleased]

## [v1.1.1] - 2021-01-06
### Updated
- Updated golangci linter to 1.33 version

### Changed
- Changed errors handling - added wrapping.

## [v1.1.0] - 2020-12-14
### Changed
- Replaced testify/assert to native tests.

## [v1.0.1] - 2020-11-14
### Added
- Added the ability to run the status command on a real Rust server. To do this, set environment variables `TEST_RUST_SERVER=true`, 
`TEST_RUST_SERVER_ADDR` and `TEST_RUST_SERVER_PASSWORD` with address and password from Rust remote console.  
- Added deadline test.  

### Changed
- Changed CI workflows and related badges. Integration with Travis-CI was changed to GitHub actions workflow. Golangci-lint 
job was joined with tests workflow.  

## [v1.0.0] - 2020-11-13
### Added
- Added mockserver and tests.
- Added `LocalAddr() net.Addr` and `RemoteAddr() net.Addr` functions that returns local and remote network addresses.

## v0.1.0 - 2020-10-22
### Added
- Initial implementation.

[Unreleased]: https://github.com/gorcon/websocket/compare/v1.1.1...HEAD
[v1.1.1]: https://github.com/gorcon/websocket/compare/v1.1.0...v1.1.1
[v1.1.0]: https://github.com/gorcon/websocket/compare/v1.0.1...v1.1.0
[v1.0.1]: https://github.com/gorcon/websocket/compare/v1.0.0...v1.0.1
[v1.0.0]: https://github.com/gorcon/websocket/compare/v0.1.0...v1.0.0
