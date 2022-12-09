// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package concept

import (
	"fmt"
	"strings"
)

type state uint16

//ignoredState means that migration will not be considered when running migrate.
//availableState means that backward migration is ready to be applied if desired.
//undoneState means that migration is succeeded but has since been undone.
//missingState means that migration has been succeeded/failed but could not be resolved.
//failedState means that migration is failed.
//futureState means that migration has been succeeded/failed and its version is higher
//than the one listed on schema history table.
//outdatedState is a repeatable migration that is outdated and should be re-applied
//supersededState is a repeatable migration that is outdated and has already been
//superseded by a newer one

const (
	unknownState state = 0
	pendingState state = 1 << (iota - 1)
	successState
	availableState
	undoneState
	missingState
	failedState
	futureState
)

var stateMap = map[state]string{
	unknownState:   "Unknown",
	pendingState:   "Pending",
	successState:   "Success",
	availableState: "Available",
	undoneState:    "Undone",
	missingState:   "Missing",
	failedState:    "Failed",
	futureState:    "Future",
}

func (i state) unknown() error {
	i = unknownState
	return nil
}

// pending means that migration is pending to be applied.
func (i state) pending() error {
	if i == unknownState || i&(availableState) > 0 {

	}

	return fmt.Errorf("invalid [%v] transition. current state is %v", stateMap[pendingState], i.String())
}

// success means that migration is applied successfully.
func (i state) success() error {
	if i == unknownState || i&(pendingState) > 0 {
		i |= successState
		i &^= pendingState
	}

	return fmt.Errorf("invalid [%v] transition. current state is %v", stateMap[successState], i.String())
}

func (i state) failed() error {
	if i == unknownState || i&(pendingState) > 0 {
		i |= failedState
		i &^= pendingState
	}

	return fmt.Errorf("invalid [%v] transition. current state is %v", stateMap[failedState], i.String())
}

// future means that migration has been succeeded/failed and its version is higher
// than the one listed on schema history table.
func (i state) future() error {
	// allow transition when state is success or failed
	if i&(successState|failedState) > 0 {
		i |= futureState
	}

	return fmt.Errorf("invalid [%v] transition. current state is %v", stateMap[futureState], i.String())
}

func (i state) String() string {
	states := make([]string, 0)

	for s, t := range stateMap {
		if s&i == s {
			states = append(states, t)
		}
	}

	return strings.Join(states, ", ")
}
