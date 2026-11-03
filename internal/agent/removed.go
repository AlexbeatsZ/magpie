package agent

import "errors"

// ErrWiringRemoved is permanent: the gateway never writes API settings into
// another program, even when an old profile still names it.
var ErrWiringRemoved = errors.New("managing other programs' API configuration has been removed; configure your client with the gateway URL and key manually")
