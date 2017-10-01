package types

// ErrRelay captures relay plugin errors
type ErrRelay struct {
	Err error
}

func (e ErrRelay) Error() string {
	return e.Err.Error()
}

// ErrSearch captures relay plugin errors
type ErrSearch struct {
	Err error
}

func (e ErrSearch) Error() string {
	return e.Err.Error()
}

// ErrRegister captures register plugin errors
type ErrRegister struct {
	Err error
}

func (e ErrRegister) Error() string {
	return e.Err.Error()
}

// ErrBookmark captures bookmark plugin errors
type ErrBookmark struct {
	Err error
}

func (e ErrBookmark) Error() string {
	return e.Err.Error()
}
