package source

import (
	"errors"
	"fmt"
	"io"
	nurl "net/url"
)

type OpenFunc func(url string) (Driver, error)

var drivers = make(map[string]OpenFunc, 0)

type Migration struct {
	Identifier string
	Script     io.ReadCloser
}

type Driver interface {
	// Name will return the driver name
	Name() string

	Close() error

	// Next prepares the next migration for reading with the Read method. It
	// returns true on success, or false if there is no next migration or an error
	// happened while preparing it. Err should be consulted to distinguish between
	// the two cases.
	//
	// Every call to Read, even the first one, must be preceded by a call to Next.
	Next() bool

	// Read the next migration. The returned migration does not guarantee
	// to be sorted. If Next is false, then calling this function will just
	// return nil.
	Read() (*Migration, error)

	// Touch will create an empty file with the given name. If the driver does not
	// support file creation, it will return error.
	Touch(name string) error

	// Remove file with the given name. If the driver does not support file deletion,
	// it will return error.
	Remove(name string) error

	// Err returns the error, if any, that was encountered during iteration.
	// Err may be called after an explicit or implicit Close.
	Err() error
}

func Open(url string) (Driver, error) {
	purl, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if purl.Scheme == "" {
		return nil, errors.New("source driver: url must include scheme as driver name")
	}

	openFunc, exists := drivers[purl.Scheme]
	if !exists {
		return nil, fmt.Errorf("source driver: unknown driver %v", purl.Scheme)
	}

	return openFunc(purl.String())
}

func Register(driver Driver, openFunc OpenFunc) {
	if drivers == nil {
		panic("source driver: driver is nil")
	}

	if openFunc == nil {
		panic("source driver: openFunc is nil")
	}

	if driver.Name() == "" {
		panic("source driver: invalid driver name")
	}

	if _, exists := drivers[driver.Name()]; exists {
		panic("source driver: driver is registered more than once")
	}

	drivers[driver.Name()] = openFunc
}
