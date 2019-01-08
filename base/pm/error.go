package pm

import (
	"fmt"
	"net/http"
)

var (
	//UnknownCommandErr          = errors.New("unknown command")
	_ RunError = &errorImpl{}
)

func UnknownCommandErr(cmd string) error {
	return &UnknownCommand{cmd}
}

type UnknownCommand struct {
	command string
}

func (e *UnknownCommand) Error() string {
	return fmt.Sprintf("unknown command: %s", e.command)
}

func IsUnknownCommand(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(*UnknownCommand)
	return ok
}

type RunError interface {
	Code() uint32
	Cause() interface{}
}

type errorImpl struct {
	code  uint32
	cause interface{}
}

func (e *errorImpl) Code() uint32 {
	return e.code
}

func (e *errorImpl) Cause() interface{} {
	return e.cause
}

func (e *errorImpl) Error() string {
	return fmt.Sprint(e.cause)
}

func Error(code uint32, cause interface{}) error {
	return &errorImpl{code: code, cause: cause}
}

func NotFoundError(cause interface{}) error {
	return Error(http.StatusNotFound, cause)
}

func BadRequestError(cause interface{}) error {
	return Error(http.StatusBadRequest, cause)
}

func ServiceUnavailableError(cause interface{}) error {
	return Error(http.StatusServiceUnavailable, cause)
}

func NotAcceptableError(cause interface{}) error {
	return Error(http.StatusNotAcceptable, cause)
}

func PreconditionFailedError(cause interface{}) error {
	return Error(http.StatusPreconditionFailed, cause)
}

func InternalError(cause interface{}) error {
	return Error(http.StatusInternalServerError, cause)
}
