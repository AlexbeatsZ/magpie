package agent

import (
	"time"
)

func aside(home string) *Agent { return asideIn(here(home)) }

func asideIn(at place) *Agent {
	c := newAsideConnection(at)
	a := &Agent{ID: "aside", Name: "Aside", Icon: "aside", Bin: "aside", Path: c.path, Dir: at.join(".aside")}

	return a
}

const asideImageKey = "imageGenerationModel"

var asideTimeout = 10 * time.Second
