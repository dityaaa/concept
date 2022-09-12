// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package migration

import (
	"github.com/dityaaa/concept/database"
	srt "sort"
	"strconv"
)

func translate(r []*database.Row) map[string]*Data {
	result := make(map[string]*Data, 0)
	for _, row := range r {
		data := &Data{
			Sequence:      &row.Sequence,
			Version:       row.Version,
			ScriptName:    row.ScriptName,
			Description:   row.Description,
			Checksum:      row.Checksum,
			AppliedBy:     row.AppliedBy,
			AppliedAt:     row.AppliedAt,
			ExecutionTime: row.ExecutionTime,
			Status:        FailedState,
		}

		if row.Success {
			data.Status = SuccessState
		}

		if _, ok := result[data.Version]; ok && row.Category == "REV" {
			data.Status |= UndoneState
		}

		result[row.Version] = data
	}
	return result
}

func sortVersion(keys []string) {
	// TODO: use natural order technique to support string based versioning (ex: V1.0, V1.2, V2.1)
	srt.SliceStable(keys, func(i, j int) bool {
		ii, _ := strconv.Atoi(keys[i])
		ij, _ := strconv.Atoi(keys[j])
		return ii < ij
	})
}

func keys[M ~map[K]V, K comparable, V any](m M) []K {
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func unique[T comparable](s []T) []T {
	values := make(map[T]bool)
	unq := make([]T, 0)
	for _, value := range s {
		if _, exists := values[value]; exists {
			continue
		}
		values[value] = true
		unq = append(unq, value)
	}

	return unq
}
