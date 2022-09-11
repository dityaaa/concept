package source

import (
	"os"
	"sort"
	"strconv"
)

type Mode string
type Origin uint8
type Meta map[string]any

const (
	AdvanceMode Mode = "ADV"
	ReverseMode Mode = "ADV"
)

const (
	UnknownOrigin Origin = iota
	LocalOrigin
	RemoteOrigin
	BothOrigin
)

type Migration struct {
	Sequence    uint64
	Version     string
	Description string
	AdvanceFile *os.File
	ReverseFile *os.File
	Checksum    string
	Meta        Meta
	State       State
}

type Migrations struct {
	indexes      []string
	migrations   map[string]*Migration
	synchronized bool
}

func NewMigration() *Migrations {
	return &Migrations{
		indexes:    make([]string, 0),
		migrations: make(map[string]*Migration),
	}
}

func (i *Migrations) Add(m *Migration) (mg *Migration, ok bool) {
	if m == nil {
		return nil, false
	}

	mg, ok = i.migrations[m.Version]
	if !ok {
		if m.State == UnknownState {
			i.synchronized = false
			m.State = PendingState
		}

		i.migrations[m.Version] = m
		i.indexes = append(i.indexes, m.Version)
		i.rebuild()
		return m, true
	}

	if m.State == mg.State {
		return nil, false
	}

	// fail if the given migration has a same mode (ADV/REV)
	if m.AdvanceFile == nil && mg.AdvanceFile == nil {
		return nil, false
	}

	notNilAssign(mg.AdvanceFile, m.AdvanceFile)
	notNilAssign(mg.ReverseFile, m.ReverseFile)

	// TODO: use FSM
	if m.ReverseFile != nil && mg.State == SuccessState {
		mg.State = AvailableState
	}

	if (m.State & (SuccessState | FailedState)) > 0 {

	}
}

func (i *Migrations) rebuild() {
	// TODO: use natural sorting algorithm
	sort.SliceStable(i.indexes, func(x, y int) bool {
		ix, _ := strconv.Atoi(i.indexes[x])
		iy, _ := strconv.Atoi(i.indexes[y])
		return ix < iy
	})

	// iterating backward to check if backward migration is possible.
	// assuming we have 5 histories, with 1, 4, and 5 being available.
	// only migration number 4 and 5 can be rolled back, because migration
	// number 2 and 3 does not available, so number 1 cannot be reached
	// for backward migration.
	for idx := len(i.indexes) - 1; idx >= 0; idx-- {

	}
}
