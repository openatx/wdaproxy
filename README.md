# wdaproxy
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=flat-square)](https://github.com/goreleaser)

Make [WebDriverAgent](https://github.com/facebook/WebDriverAgent) more powerful.

# Platform
Limited in Mac

# Features
- [x] Launch iproxy when start
- [x] Create http proxy for WDA server
- [x] add udid into `GET /status`
- [x] forward all url starts with `/origin/<url>` to `/<url>`
- [x] Add the missing Index page
- [x] Support Package management API
- [x] Support launch WDA
- [x] iOS device remote control

# Installl
```
$ brew install openatx/tap/wdaproxy
```

# Usage
Simple run 

```
$ wdaproxy -p 8100 -u $UDID
```

Run with WDA

```
$ wdaproxy -W ../WebDriverAgent
```

For more run `wdaproxy -h`

# Extended API
Package install

```
$ curl -X POST -F file=@some.ipa http://localhost:8100/api/v1/packages
$ curl -X POST -F url="http://somehost.com/some.ipa" http://localhost:8100/api/v1/packages
```

Package uninstall

```
$ curl -X DELETE http://localhost:8100/api/v1/packages/${BUNDLE_ID}
```

Package list (parse from `ideviceinstaller -l`)

```
$ curl -X GET http://localhost:8100/api/v1/packages
```

# For developer
Update golang vendor
```
brew install glide
glide up
```

# LICENSE
Under [MIT](LICENSE)
