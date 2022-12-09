package database

import (
	"errors"
	"fmt"
	"io"
	nurl "net/url"
)

type OpenFunc func(url string) (Driver, error)

var drivers = make(map[string]OpenFunc, 0)

type History struct {
	Rank          uint64
	Mode          string
	Version       string
	ScriptName    string
	Description   string
	Checksum      string
	AppliedBy     string
	AppliedAt     uint64
	ExecutionTime uint32
	Success       bool
}

type Driver interface {
	// Name will return the driver name
	Name() string

	Close() error
	Read() ([]*History, error)
	Write(*History) error
	Run(migration io.Reader) error
	Purge() []error
}

type Locker interface {
	Lock()
	Unlock()

	// Locked returns current shared lock status
	Locked() bool

	// Lockable returns true if shared lock is enabled
	Lockable() bool
}

func Open(url string) (Driver, error) {
	purl, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if purl.Scheme == "" {
		return nil, errors.New("database driver: url must include scheme as driver name")
	}

	if !purl.Query().Has("x-history-table") {
		purl.Query().Set("x-history-table", "migration_history")
	}

	if !purl.Query().Has("x-locking-table") {
		purl.Query().Set("x-locking-table", "migration_locking")
	}

	if !purl.Query().Has("x-without-locking") {
		purl.Query().Del("x-locking-table")
	}

	openFunc, exists := drivers[purl.Scheme]
	if !exists {
		return nil, fmt.Errorf("database driver: unknown driver %v", purl.Scheme)
	}

	return openFunc(purl.String())
}

func Register(driver Driver, openFunc OpenFunc) {
	if drivers == nil {
		panic("database driver: driver is nil")
	}

	if openFunc == nil {
		panic("database driver: openFunc is nil")
	}

	if driver.Name() == "" {
		panic("database driver: invalid driver name")
	}

	if _, exists := drivers[driver.Name()]; exists {
		panic("database driver: driver is registered more than once")
	}

	drivers[driver.Name()] = openFunc
}
