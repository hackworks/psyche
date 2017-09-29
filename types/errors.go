package types

// ErrRelay captures relay plugin errors
type ErrRelay struct {
	Err error
}

func (e ErrRelay) Error() string {
	return e.Err.Error()
}

// ErrRegister captures register plugin errors
type ErrRegister struct {
	Err error
}

func (e ErrRegister) Error() string {
	return e.Err.Error()
}
