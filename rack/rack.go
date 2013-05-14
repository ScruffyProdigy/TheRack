/*
	The Rack package allows you to break down a task into a series of smaller functions (called Middleware)

	Each Middleware gets inserted into a Rack, and then rack will proceed to call them one by one.
	Each one returns the combined results to the previous one, which can then be adjusted or altered

	This ends up being very useful in conjuction with httper for websites, 
	as a easier alternative to using the provided http.ResponseWriter.
	The rack allows us to abstract away different parts of the program,
	which allows us to easily reuse smaller parts of the program in new websites.
	Also, we now have the ability to make alter our response:
	Once something has been written to the ResponseWriter, there's no way to undo it, or to even change the headers.
	Middleware will frequently change the responses handed down by later Middleware, or write headers even after we know what most of the response will be
*/
package rack

//A Connection is something that uses a rack to complete a task, for example an HTTP connection will use a rack to return an HTTP response to an HTTP request
type Connection interface {
	Go(Middleware) error
}

// Middleware is the interface that we use to allow different pieces of your web server to communicate with one another
// typically previous middleware will take the results and manipulate them as necessary
type Middleware interface {
	// Run takes a map of all of the variables, and a way to access the next part of the server
	Run(vars map[string]interface{}, next func())
}

// an ordinary function can satisfy the middleware interface, if you cast it into a rack.Func
type Func func(vars map[string]interface{}, next func())

func (this Func) Run(vars map[string]interface{}, next func()) {
	this(vars, next)
}

// Rack is a collection of middleware that is also a middleware itself
// when run, it simply runs through each of the middleware within it, passing each one a link to the next one
type Rack []Middleware

// New is a utility function to get a basic rack
// By default, it assumes that you want at least 2 middleware within it, and makes enough room for them
// it will accomodate as many as you need, though
func New() *Rack {
	rack := make(Rack, 0, 2)
	return &rack
}

// Add will put another middleware into the list
// it is order dependent, so make sure any dependent middleware are put in after any required middleware
func (this *Rack) Add(m Middleware) {
	*this = append(*this, m)
}

func (this Rack) Run(vars map[string]interface{}, next func()) {
	index := -1
	var ourNext func()
	ourNext = func() {
		index++
		defer func() { index-- }() //if anything branches, we want the index to sync up correctly

		if index >= len(this) {
			next()
		} else {
			this[index].Run(vars, ourNext)
		}
	}
	ourNext()
}

//NewVars creates a new set of variables
func NewVars() map[string]interface{} {
	return make(map[string]interface{})
}
