package errs

import (
	"errors"
)

var ErrInternalSrv = errors.New("internal error on server")           //Internal error
var ErrCorrectURL = errors.New("enter correct url parameter")         //Enter correct url
var ErrNoContent = errors.New("no content")                           //No content
var ErrBadRequest = errors.New("bad request")                         //Bad request
var ErrEmptyBody = errors.New("body is empty")                        //Body empty
var ErrFieldsJSON = errors.New("invalid fields in json")              //Invalid fields is json
var ErrURLNotFound = errors.New("url not found")                      //URL not found
var ErrAlreadyHasShort = errors.New("already has short")              //Already has short link
var ErrURLIsGone = errors.New("url is gone")                          //URL is gone
var ErrUnknownEnvOrFlag = errors.New("unknown env or flag param")     //Unknown env of flag
var ErrFileStorageNotClose = errors.New("file storage has not close") //File storage not close
var ErrCorrelation = errors.New(("correlationIds is null"))           //CorrelationIds is null
