// +build dev

package main

import "net/http"

// Assets contains project assets.
var Assets http.FileSystem = http.Dir("assets")
