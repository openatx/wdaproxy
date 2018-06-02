# wdaproxy
[![Powered By: GoReleaser](https://img.shields.io/badge/powered%20by-goreleaser-green.svg?style=flat-square)](https://github.com/goreleaser)

Make [WebDriverAgent](https://github.com/facebook/WebDriverAgent) more powerful.

# Platform
Limited in Mac

# Features
- [x] Launch iproxy when start. Listen on `0.0.0.0` instead of `localhost`
- [x] Create http proxy for WDA server
- [x] add udid into `GET /status`
- [x] forward all url starts with `/origin/<url>` to `/<url>`
- [x] Add the missing Index page
- [x] Support Package management API
- [x] Support launch WDA
- [x] iOS device remote control
- [x] Support Appium Desktop (Beta)

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

Strong recommended use [python facebook-wda](https://github.com/openatx/facebook-wda) to write tests.
But if you have to use appium. Just keep reading.

## Simulate appium server
Since WDA implements WebDriver protocol. 
Even through many API not implemented. But it still works. `wdaproxy` implemented a few api listed bellow.

- wdaproxy "/wd/hubs/sessions"
- wdaproxy "/wd/hubs/session/$sessionId/window/current/size"

Launch wdaproxy with command `wdaproxy -p 8100 -u $UDID`

Here is sample `Python-Appium-Client` code.

```python
from appium import webdriver
import time

driver = webdriver.Remote("http://127.0.0.1:8100/wd/hub", {"bundleId": "com.apple.Preferences"})

def wait_element(xpath, timeout=10):
    print("Wait ELEMENT", xpath)
    deadline = time.time() + timeout
    while time.time() <= deadline:
        el = driver.find_element_by_xpath(xpath)
        if el:
            return el
        time.sleep(.2)
    raise RuntimeError("Element for " + xpath + " not found")

wait_element('//XCUIElementTypeCell[@name="蓝牙"]').click()
```

Not working well code

```python
driver.background_app(3)
driver.implicitly_wait(30)
driver.get_window_rect()
# many a lot.
```

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
First checkout this repository

```bash
git clone https://github.com/openatx/wdaproxy $GOPATH/src/github.com/openatx/wdaproxy
cd $GOPATH/src/github.com/openatx/wdaproxy
```

Update golang vendor
```bash
brew install glide
glide up
```

Package web resources into binary

```bash
go generate ./web
go build -tags vfs
```

# LICENSE
Under [MIT](LICENSE)
