package rack

// the "Vars" is just a way of holding static variables so that one piece of middleware can access results from another piece of middleware
type Vars map[string]interface{}

// a utility function to create a place to store the variables
func NewVars() Vars {
	return make(Vars)
}

// You can also use a Vars as a piece of Middleware itself, and it will just set default variables for later pieces of middleware
func (this Vars) Run(vars Vars, next func()) {
	for k, v := range this {
		vars[k] = v
	}
	next()
}

func (this Vars) Set(k string, v interface{}) {
	this[k] = v
}

func (this Vars) Get(k string) interface{} {
	return this[k]
}

func (this Vars) SetIfEmpty(k string, v interface{}) {
	if this[k] == nil {
		this[k] = v
	}
}

func (this Vars) Clear(k string) (result interface{}) {
	result = this[k]
	delete(this, k)
	return
}

func (this Vars) Switch(k string, v interface{}) (result interface{}) {
	result = this[k]
	this[k] = v
	return
}
