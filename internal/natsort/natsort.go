// Package natsort implements natural strings sorting
//
// Copyright (c) 2015, Vincent Batoufflet and Marc Falzon
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
//
// * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright
// notice, this list of conditions and the following disclaimer in the
// documentation and/or other materials provided with the distribution.
//
// * Neither the name of the authors nor the names of its contributors
// may be used to endorse or promote products derived from this software
// without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.
package natsort

import (
	"sort"
	"strings"
)

type stringSlice []string

func (s stringSlice) Len() int {
	return len(s)
}

func (s stringSlice) Less(a, b int) bool {
	return Compare(s[a], s[b])
}

func (s stringSlice) Swap(a, b int) {
	s[a], s[b] = s[b], s[a]
}

// Sort sorts a list of strings in a natural order
func Sort(l []string) {
	sort.Sort(stringSlice(l))
}

// Compare returns true if the first string precedes the second one according to natural order
func Compare(a, b string) bool {
	ln_a := len(a)
	ln_b := len(b)
	posa := 0
	posb := 0

	for {
		if ln_a <= posa {
			if ln_b <= posb {
				// eof on both at the same time (equal)
				return false
			}
			return true
		} else if ln_b <= posb {
			// eof on b
			return false
		}

		av, bv := a[posa], b[posb]

		if av >= '0' && av <= '9' && bv >= '0' && bv <= '9' {
			// go into numeric mode
			intlna := 1
			intlnb := 1
			for {
				if posa+intlna >= ln_a {
					break
				}
				x := a[posa+intlna]
				if av == '0' {
					posa += 1
					av = x
					continue
				}
				if x >= '0' && x <= '9' {
					intlna += 1
				} else {
					break
				}
			}
			for {
				if posb+intlnb >= ln_b {
					break
				}
				x := b[posb+intlnb]
				if bv == '0' {
					posb += 1
					bv = x
					continue
				}
				if x >= '0' && x <= '9' {
					intlnb += 1
				} else {
					break
				}
			}
			if intlnb > intlna {
				// length of a value is longer, means it's a bigger number
				return true
			} else if intlna > intlnb {
				return false
			}
			// both have same length, let's compare as string
			v := strings.Compare(a[posa:posa+intlna], b[posb:posb+intlnb])
			if v < 0 {
				return true
			} else if v > 0 {
				return false
			}
			// equale
			posa += intlna
			posb += intlnb
			continue
		}

		if av == bv {
			posa += 1
			posb += 1
			continue
		}

		return av < bv
	}
	return false
}
