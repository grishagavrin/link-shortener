# run macOS

1.run in cmd/shortener: nodemon --watch ../../ --exec go run main.go --signal SIGTERM

# run wnd

1.run in cmd/shortener: nodemon --watch ../../ --exec go run main.go --signal SIGKILL

# run tests

1. run tests in root directory of project go test ./... -v

# test db

1. postgresql://postgres:**\*\***@127.0.0.1:5432/golangDB

# bench tests

1. go test -bench=BenchmarkHandler_SaveTXT -benchmem -benchtime=2500x -memprofile base.pprof // Run one bench
2. go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof // See difference
3. go test -bench=. -memprofile=base.pprof // Run all benchs in handlers
4. go tool pprof -http=":9090" handlers.test base.pprof // See profile
5. go tool pprof -http=":9090" handlers.test result.pprof // See result profile

# go fmt

1. gofmt -s -w . in root directory
