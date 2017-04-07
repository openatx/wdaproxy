# muuid
[![GoDoc](https://godoc.org/github.com/codeskyblue/muuid?status.svg)](https://godoc.org/github.com/codeskyblue/muuid)

Machine UUID, port of github.com/mhzed/machine-uuid

## Install
```
go get -v github.com/codeskyblue/muuid
```

## Usage
```go
package main

import "github.com/codeskyblue/muuid"

func main(){
	println(muuid.UUID()) // same as muuid.UUIDFromOS(runtime.GOOS)

	// Generate UUID and put to ~/.muid
    // Not read from /var/lib/dbus/machine-id, for RPi image matchine-id are all the same.
	println(muuid.UUIDFromOS("raspberry")) 
}
```

## LICENSE
[MIT](LICENSE)
