// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migration

type State uint16

const (
	UnknownState State = 0

	//PendingState means that migration has not been applied yet.
	PendingState State = 1 << (iota - 1)

	//SuccessState means that migration is succeeded.
	SuccessState

	//IgnoredState means that migration will not be considered when running migrate.
	IgnoredState

	//AvailableState means that backward migration is ready to be applied if desired.
	AvailableState

	//UndoneState means that migration is succeeded but has since been undone.
	UndoneState

	//MissingState means that migration has been succeeded/failed but could not be resolved.
	MissingState

	//FailedState means that migration is failed.
	FailedState

	//FutureState means that migration has been succeeded/failed and its version is higher
	//than the one listed on schema history table.
	FutureState

	//OutdatedState is a repeatable migration that is outdated and should be re-applied
	OutdatedState

	//SupersededState is a repeatable migration that is outdated and has already been
	//superseded by a newer one
	SupersededState
)

// TODO: complete state transition map
var stateMap = map[State]map[State]struct{}{
	UnknownState: {
		UnknownState:    {},
		PendingState:    {},
		SuccessState:    {},
		IgnoredState:    {},
		AvailableState:  {},
		UndoneState:     {},
		MissingState:    {},
		FailedState:     {},
		FutureState:     {},
		OutdatedState:   {},
		SupersededState: {},
	},
	PendingState: {
		PendingState:   {},
		SuccessState:   {},
		AvailableState: {},
	},
	UndoneState: {
		UndoneState:  {},
		SuccessState: {},
		FailedState:  {},
	},
}

func transitionAllowed(from, to State) bool {
	f, fOk := stateMap[from]
	if !fOk {
		return false
	}

	_, tOk := f[to]
	return tOk
}

func TranslateState(state State) []string {
	states := make([]string, 0)

	if state == UnknownState {
		return append(states, "Unknown")
	}

	if (state & PendingState) > 0 {
		states = append(states, "Pending")
	}

	if (state & SuccessState) > 0 {
		states = append(states, "Success")
	}

	if (state & IgnoredState) > 0 {
		states = append(states, "Ignored")
	}

	if (state & AvailableState) > 0 {
		states = append(states, "Available")
	}

	if (state & UndoneState) > 0 {
		states = append(states, "Undone")
	}

	if (state & MissingState) > 0 {
		states = append(states, "Missing")
	}

	if (state & FailedState) > 0 {
		states = append(states, "Failed")
	}

	if (state & FutureState) > 0 {
		states = append(states, "Future")
	}

	if (state & OutdatedState) > 0 {
		states = append(states, "Outdated")
	}

	if (state & SupersededState) > 0 {
		states = append(states, "Superseded")
	}

	return states
}
