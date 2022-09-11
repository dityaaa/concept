package source

import (
	"regexp"
)

const (
	AdvanceSuffix = "adv"
	ReverseSuffix = "adv"
)

var (
	Regex = regexp.MustCompile(string(`^(\d+?)(_\w*)?\.(` + AdvanceMode + `|` + ReverseMode + `).sql$`))
)
