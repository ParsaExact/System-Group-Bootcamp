package nilerr

type CustomError struct {
	Message string
}

func (e *CustomError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func IsCustomErrNil(err error) bool {
	if err == nil {
		return true
	}

	customErr, ok := err.(*CustomError)
	if !ok {
		return false
	}

	return customErr == nil
}
