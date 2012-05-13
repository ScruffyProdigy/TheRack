/*
	The Rack package allows you to break down your web server into smaller functions (called Middleware)

	Each Middleware gets inserted into a Rack, and then rack will proceed to call them one by one.
	Each one returns the combined results to the previous one, which can then be adjusted or altered

	This works as an alternative to simply writing to the provided http.ResponseWriter
	The advantage to using this method is an easy way to abstract away different parts of the program,
	which allows us to easily reuse smaller parts of the program in new websites.
	Also, we now have the ability to make adjustments to other parts of the program:
	Once something has been written to the ResponseWriter, there's no way to undo it, or to even change the headers.
	Middleware will frequently change the responses handed down by later Middleware, or write headers even after we know what most of the response will be
*/
package rack

//A Connection is something that uses a rack to complete a task, for example an HTTP connection will use a rack to return an HTTP response to an HTTP request
type Connection interface {
	Go(Middleware) error
}

// the "Middleware" is the interface that we use to allow different pieces of your web server to communicate with one another
// "Run" takes in the http Request that we've received, a map of all of the variables, and a way to access the next part of the server
// it returns the http Status, the headers, and a byte encoded response
// typically previous middleware will take the results and manipulate them as necessary
type Middleware interface {
	Run(vars Vars, next func())
}

// an ordinary function can satisfy the middleware interface, if you cast it into a rack.Func
type Func func(vars Vars, next func())

func (this Func) Run(vars Vars, next func()) {
	this(vars, next)
}

// a "Rack" is a collection of middleware that is also a middleware itself
// when run, it simply runs through each of the middleware within it, passing each one a link to the next one
type Rack []Middleware

// "New" is a utility function to get a basic rack
// By default, it assumes that you want at least 2 middleware within it, and makes enough room for them
// it will accomodate as many as you need, though
func New() *Rack {
	rack := make(Rack, 0, 2)
	return &rack
}

// "Add" will put another middleware into the list
// it is order dependent, so make sure any dependent middleware are put in after any required middleware
func (this *Rack) Add(m Middleware) {
	*this = append(*this, m)
}

func (this Rack) Run(vars Vars, next func()) {
	index := -1
	var ourNext func()
	ourNext = func() {
		index++
		if index >= len(this) {
			next()
		} else {
			this[index].Run(vars, ourNext)
		}
		index--
	}
	ourNext()
}
