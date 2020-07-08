package msmodule

import (
	"fmt"
	"regexp"
)

var (
	_ MsRouter = (*msRouter)(nil)

	isAlphaNumeric = regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString
)

// MsRouter defines multi signature message router (copy from gov module).
type MsRouter interface {
	AddRoute(r string, h MsHandler) MsRouter
	HasRoute(r string) bool
	GetRoute(path string) MsHandler
	Seal()
}

// msRouter is a MsRouter implementation.
type msRouter struct {
	routes map[string]MsHandler
	sealed bool
}

// Seal seals the router which prohibits any subsequent route handlers to be
// added. Seal will panic if called more than once.
func (rtr *msRouter) Seal() {
	if rtr.sealed {
		panic("router already sealed")
	}
	rtr.sealed = true
}

// AddRoute adds a governance handler for a given path. It returns the Router
// so AddRoute calls can be linked. It will panic if the router is sealed.
func (rtr *msRouter) AddRoute(path string, h MsHandler) MsRouter {
	if rtr.sealed {
		panic("router sealed; cannot add route handler")
	}
	if !isAlphaNumeric(path) {
		panic("route expressions can only contain alphanumeric characters")
	}
	if rtr.HasRoute(path) {
		panic(fmt.Sprintf("route %s has already been initialized", path))
	}

	rtr.routes[path] = h

	return rtr
}

// HasRoute returns true if the router has a path registered or false otherwise.
func (rtr *msRouter) HasRoute(path string) bool {
	return rtr.routes[path] != nil
}

// GetRoute returns a Handler for a given path.
func (rtr *msRouter) GetRoute(path string) MsHandler {
	if !rtr.HasRoute(path) {
		panic(fmt.Sprintf("route \"%s\" does not exist", path))
	}

	return rtr.routes[path]
}

// NewMsRouter creates a new multi signature router.
func NewMsRouter() MsRouter {
	return &msRouter{
		routes: make(map[string]MsHandler),
	}
}
