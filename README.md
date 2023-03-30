# run macOS

1. nodemon --watch ../../ --exec go run main.go --signal SIGTERM

# run wnd

1. nodemon --watch ../../ --exec go run main.go --signal SIGKILL

# run tests

1. run tests in root directory of project go test ./... -v
