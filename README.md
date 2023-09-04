# run macOS

in cmd/shortener

    nodemon --watch ../../ --exec go run main.go --signal SIGTERM

# run wnd

in cmd/shortener

    nodemon --watch ../../ --exec go run main.go --signal SIGKILL

# run tests

in root directory

    run tests in root directory of project go test ./... -v

# test db

in internal/config DatabaseDSN change envDefault

    postgresql://postgres:***@127.0.0.1:5432/golangDB

# bench tests

in internal/handlers/bench

    go test -bench=. -benchmem -benchtime=10000x -memprofile base.pprof

show pprof

    go tool pprof -http=":9090" bench.test base.pprof

show difference

    go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof
