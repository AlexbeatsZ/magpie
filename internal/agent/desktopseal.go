package agent

import (
	"errors"
)

// errDesktopSealed: Desktop sealed a value with a keystore magpie can't
// open here.
var errDesktopSealed = errors.New("sealed by Claude Desktop")
