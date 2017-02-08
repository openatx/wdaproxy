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

# Usage
```
$ brew install openatx/tap/wdaproxy
$ wdaproxy -p 8100 -u $UDID
```


# LICENSE
Under [MIT](LICENSE)
