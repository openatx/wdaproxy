# Interface
API which offer to FE

## File operation
Get file list

```
$ curl -X GET /api/v1/files
{
	"success": true,
	"value": [{
		"name": "a/f.txt"
	}, {
		"name": "b/f.py"
	}]
}
```

Delete file

```
$ curl -X DELETE /api/v1/files/{name}
{
	"success": true,
	"description": "file deleted"
}
```

Add file

```
$ curl -X POST /api/v1/files <<EOF
{
	"name": "some.png",
	"data": "# coding: utf-8"
}
EOF
{
	"success": true,
	"description": "file created"
}
```

## Devices
```
$ curl -X GET /api/v1/devices
{
	"success": true,
	"value": [{
		"platform": "Android",
		"serial": "XAFEFF"
	}, {
		"platform": "iOS",
		"udid": "11231lk2j3lkjsdfasdfafaff"
	}]
}
```

## Screenshot
```
$ curl -X GET /api/v1/devices/{serial}/screenshot
{
	"success": true,
	"imageType": "png",
	"value": "LKJ#RF"
}
```

## Code debug
WebSocket 

```
ws.connect ws://hostname.com/ipython

ws.send {
	"code": "# coding: utf-8 ..."
}

ws.recv {
	"output": "hello\n"
}

in the last recv: {
	"elapsedTime": 10002, // unit ms
	"exitCode": 0
}
```