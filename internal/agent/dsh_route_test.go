package agent

import (
	"testing"
)

func TestDshSaidOnceSaysATroubleOnce(t *testing.T) {
	var said dshSaidOnce
	const trouble = "writing dsh's route again in /x: open: permission denied"
	if !said.first(trouble) {
		t.Fatal("the first round said nothing")
	}
	if said.first(trouble) {
		t.Fatal("the same trouble was said again")
	}
	if said.first("") {
		t.Fatal("a quiet round said something")
	}
	if !said.first(trouble) {
		t.Fatal("trouble after a quiet round was not said")
	}
}
