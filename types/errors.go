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

// ErrIndexer captures indexer plugin errors
type ErrIndexer struct {
	Err error
}

func (e ErrIndexer) Error() string {
	return e.Err.Error()
}
