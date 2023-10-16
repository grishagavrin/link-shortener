// Package errs consist package level errors
package errs

import (
	"errors"
)

// HTTP Errors

// Internal error
var ErrInternalSrv = errors.New("internal error on server")

// Enter correct url
var ErrCorrectURL = errors.New("enter correct url parameter")

// No content
var ErrNoContent = errors.New("no content")

// Bad request
var ErrBadRequest = errors.New("bad request")

// Body empty
var ErrEmptyBody = errors.New("body is empty")

// URL not found
var ErrURLNotFound = errors.New("url not found")

// Already has short link
var ErrAlreadyHasShort = errors.New("already has short")

// URL is gone
var ErrURLIsGone = errors.New("url is gone")

// File storage not close
var ErrFileStorageNotClose = errors.New("file storage has not close")

// CorrelationIds is null
var ErrCorrelation = errors.New("correlationIds is null")

// Read body error
var ErrReadAll = errors.New("something went wrong with read body")

// URL not found in DB
var ErrNotFoundURL = errors.New("url not found in DB")

// DB not avaliable
var ErrDatabaseNotAvaliable = errors.New("db not avaliable")

// Exec error DB
var ErrDatabaseExec = errors.New("exec query error in db")

// Query error DB
var ErrDatabaseQuery = errors.New("query error in db")

// Scan rows error DB
var ErrDatabaseScanRows = errors.New("scan rows error in db")

// RAM not avaliable
var ErrRAMNotAvaliable = errors.New("ram not avaliable")

// Initialize error logger
var ErrInitLogger = errors.New("can`t initialize logger")

// Invalid fields is json
var ErrFieldsJSON = errors.New("invalid fields in json")

// Can`t unmarshall
var ErrJSONUnMarshall = errors.New("cant unmarshall")

// Can`t marshall
var ErrJSONMarshall = errors.New("cant marshall")

// Unknown env of flag
var ErrUnknownEnvOrFlag = errors.New("unknown env or flag param")

// Can`t load ENV
var ErrENVLoading = errors.New("can`t load ENV")

// Config instance error
var ErrConfigInstance = errors.New("get config instance error")

// Config value error
var ErrConfigValue = errors.New("get config value error: ")
