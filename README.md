# run HTTPS

in cmd/shortener

    go run main.go -s 'ssl'

# run macOS

in cmd/shortener

    nodemon --watch ./ --exec go run ./cmd/shortener/main.go --signal SIGTERM

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

go fmt before commit in root dir

    go fmt ./...

# get godoc

in root dir

    godoc -http=:9090 and tap in browser http://localhost:9090/pkg/?m=all

# generate swagger

in root dir

    swag init -g .\cmd\shortener\main.go

# gRPC

in internal/proto dir

    protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    ./link_shortener.proto
