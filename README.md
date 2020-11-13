# Websocket
[![GitHub Build](https://github.com/gorcon/websocket/workflows/build/badge.svg)](https://github.com/gorcon/websocket/actions?query=workflow%3Abuild)
[![Coverage](https://gocover.io/_badge/github.com/gorcon/websocket?0 "coverage")](https://gocover.io/github.com/gorcon/websocket)
[![Go Report Card](https://goreportcard.com/badge/github.com/gorcon/websocket)](https://goreportcard.com/report/github.com/gorcon/websocket)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/gorcon/websocket)

Rust Web RCON Protocol implementation in Go.

## Supported Games
* [Rust](https://store.steampowered.com/app/252490) (add +rcon.web 1 to the args when starting the server)

## Install
```text
go get github.com/gorcon/websocket
```

See [Changelog](CHANGELOG.md) for release details.

## Usage
```go
package main

import (
	"log"
	"fmt"

	"github.com/gorcon/websocket"
)

func main() {
	conn, err := websocket.Dial("127.0.0.1:28016", "password")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	response, err := conn.Execute("status")
	if err != nil {
		log.Fatal(err)
	}
	
	fmt.Println(response)	
}
```

## Requirements
Go 1.15 or higher

## Contribute
Contributions are more than welcome! 

If you think that you have found a bug, create an issue and publish the minimum amount of code triggering the bug so 
it can be reproduced.

If you want to fix the bug then you can create a pull request. If possible, write a test that will cover this bug.

## License
MIT License, see [LICENSE](LICENSE)
