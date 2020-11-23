module github.com/openatx/wdaproxy

go 1.13

require (
	github.com/codeskyblue/muuid v0.0.0-20170401091614-44f8dfd4b3a9
	github.com/facebookgo/freeport v0.0.0-20150612182905-d4adf43b75b9
	github.com/gobuild/log v1.0.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/mash/go-accesslog v1.2.0
	github.com/ogier/pflag v0.0.1
	github.com/pkg/errors v0.9.1
	github.com/satori/go.uuid v1.2.0 // indirect
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20200824052919-0d455de96546
	github.com/stretchr/testify v1.6.1 // indirect
	golang.org/x/tools v0.0.0-20201121010211-780cb80bd7fb // indirect
	howett.net/plist v0.0.0-20201026045517-117a925f2150
)

replace github.com/qiniu/log => github.com/gobuild/log v0.1.0
