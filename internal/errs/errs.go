package errs

import (
	"errors"
)

var ErrInternalSrv = errors.New("internal error on server")   //Internal error
var ErrCorrectURL = errors.New("enter correct url parameter") //Enter correct url
var ErrNoContent = errors.New("no content")                   //No content
var ErrBadRequest = errors.New("bad request")                 //Bad request
var ErrEmptyBody = errors.New("body is empty")                //Body empty
var ErrFieldsJSON = errors.New("invalid fields in json")      //Invalid fields is json
