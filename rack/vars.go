package rack

import (
	"net/http"
)

// the "Vars" is just a way of holding static variables so that one piece of middleware can access results from another piece of middleware
type Vars map[string]interface{}

// a utility function to create a place to store the variables
func NewVars() Vars {
	return make(Vars)
}

// You can also use a Vars as a piece of Middleware itself, and it will just set default variables for later pieces of middleware
func (this Vars) Run(r *http.Request, vars Vars, next Next) (status int, header http.Header, message []byte) {
	for k, v := range this {
		vars[k] = v
	}
	return next()
}

// in order to make it easier to do stuff to the vars, you can create a VarFunc, and call Apply
// this makes it easier to make shortcuts for commonly applied Vars routines
type VarFunc func(Vars) interface{}

func (this Vars) Apply(f VarFunc) interface{} {
	return f(this)
}
