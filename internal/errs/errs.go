package errs

import (
	"errors"
)

// HTTP Errors
var ErrInternalSrv = errors.New("internal error on server")           //Internal error
var ErrCorrectURL = errors.New("enter correct url parameter")         //Enter correct url
var ErrNoContent = errors.New("no content")                           //No content
var ErrBadRequest = errors.New("bad request")                         //Bad request
var ErrEmptyBody = errors.New("body is empty")                        //Body empty
var ErrURLNotFound = errors.New("url not found")                      //URL not found
var ErrAlreadyHasShort = errors.New("already has short")              //Already has short link
var ErrURLIsGone = errors.New("url is gone")                          //URL is gone
var ErrFileStorageNotClose = errors.New("file storage has not close") //File storage not close
var ErrCorrelation = errors.New("correlationIds is null")             //CorrelationIds is null
var ErrReadAll = errors.New("something went wrong with read body")    //Read body error
var ErrNotFoundURL = errors.New("url not found in DB")

// DB Storage Errors
var ErrDatabaseNotAvaliable = errors.New("db not avaliable")  // DB not avaliable
var ErrDatabaseExec = errors.New("exec query error in db")    // Exec error DB
var ErrDatabaseQuery = errors.New("query error in db")        // Query error DB
var ErrDatabaseScanRows = errors.New("scan rows error in db") // Scan rows error DB
// RAM Storage Storage
var ErrRAMNotAvaliable = errors.New("ram not avaliable") // RAM not avaliable
// Logger errors
var ErrInitLogger = errors.New("can`t initialize logger") // Initialize error logger
// JSON
var ErrFieldsJSON = errors.New("invalid fields in json") //Invalid fields is json
var ErrJSONUnMarshall = errors.New("cant unmarshall")    // Can`t unmarshall
var ErrJSONMarshall = errors.New("cant marshall")        // Can`t marshall
// ENV
var ErrUnknownEnvOrFlag = errors.New("unknown env or flag param") //Unknown env of flag
