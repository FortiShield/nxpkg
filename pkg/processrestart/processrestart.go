package processrestart

import "errors"

// CanRestart reports whether the current set of Nxpkg processes can
// be restarted.
func CanRestart() bool {
	return usingGoremanDev || usingGoremanServer
}

// Restart restarts the current set of Nxpkg processes associated with
// this server.
func Restart() error {
	if !CanRestart() {
		return errors.New("reloading site is not supported")
	}
	if usingGoremanDev {
		return restartGoremanDev()
	}
	if usingGoremanServer {
		return restartGoremanServer()
	}
	return errors.New("unable to restart processes")
}

// WillRestart is a channel that is closed when the process will imminently restart.
var WillRestart = make(chan struct{})
