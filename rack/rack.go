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

import (
	"net/http"
)

// a "Next" is a function that gets passed to each of the Middleware so that they can interact with the next piece of Middleware
type Next func() (int, http.Header, []byte)

// the "Middleware" is the interface that we use to allow different pieces of your web server to communicate with one another
// "Run" takes in the http Request that we've received, a map of all of the variables, and a way to access the next part of the server
// it returns the http Status, the headers, and a byte encoded response
// typically previous middleware will take the results and manipulate them as necessary
type Middleware interface {
	Run(req *http.Request, vars Vars, next Next) (status int, header http.Header, message []byte)
}

// an ordinary function can satisfy the middleware interface, if you cast it into a rack.Func
type Func func(req *http.Request, vars Vars, next Next) (status int, header http.Header, message []byte)

func (this Func) Run(req *http.Request, vars Vars, next Next) (status int, header http.Header, message []byte) {
	return this(req, vars, next)
}

// a "Rack" is a collection of middleware that is also a middleware itself
// when run, it simply runs through each of the middleware within it, passing each one a link to the next one
type Rack []Middleware

// "NewRack" is a utility function to get a basic rack
// By default, it assumes that you want at least 2 middleware within it, and makes enough room for them
// it will accomodate as many as you need, though
func NewRack() *Rack {
	rack := make(Rack, 0, 2)
	return &rack
}

// "Add" will put another middleware into the list
// it is order dependent, so make sure any dependent middleware are put in after any required middleware
func (this *Rack) Add(m Middleware) {
	*this = append(*this, m)
}

func (this Rack) Run(r *http.Request, vars Vars, next Next) (status int, header http.Header, message []byte) {
	index := -1
	var ourNext Next
	ourNext = func() (int, http.Header, []byte) {
		index++
		if index >= len(this) {
			return next()
		}
		return this[index].Run(r, vars, ourNext)
	}
	return ourNext()
}

//Up is the public interface to the default Rack
var Up *Rack = NewRack()

// "NotFound" is the default next function
// we assume that if the final middleware calls next(), that none of the middleware were fully able to handle the request
func NotFound() (status int, header http.Header, message []byte) {
	return http.StatusNotFound, make(http.Header), []byte("")
}

// "Run" tells a connection to run a specific middleware
func Run(c Connection, m Middleware) error {
	return c.Go(func(w http.ResponseWriter, r *http.Request) {
		vars := NewVars()
		status, headers, message := m.Run(r, vars, NotFound)
		for k, _ := range headers {
			w.Header().Set(k, headers.Get(k))
		}
		w.WriteHeader(status)
		w.Write(message)
	})
}