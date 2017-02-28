# wdaproxy
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=flat-square)](https://github.com/goreleaser)
WebDriverAgent Proxy

Only can work in Mac

# Features
- [x] Launch iproxy when start
- [x] Create http proxy for WDA server
- [x] add udid into `GET /status`
- [x] forward all url starts with `/origin/<url>` to `/<url>`
- [x] Add the missing Index page
- [x] Support URL `GET /packages` and `DELETE /packages/{bundleId}`
- [x] Support launch WDA

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

# LICENSE
Under [MIT](LICENSE)
