# Rack

A simple rack implementation in Go

run `go get github.com/HairyMezican/The-Rack/rack` to install

## Examples

	package main

	import (
	    "github.com/HairyMezican/The-Rack/rack"
	    "net/http"
	)

	type HelloWare struct{}

	func (HelloWare) Run(r *http.Request, vars rack.Vars, next rack.Next) (int, http.Header, []byte) {
	    return http.StatusOK, make(http.Header), []byte("Hello " + vars["Object"].(string))
	}

	type WorldWare struct{}

	func (WorldWare) Run(r *http.Request, vars rack.Vars, next rack.Next) (int, http.Header, []byte) {
	    vars["Object"] = "World"
	    return next()
	}

	func main() {
	    conn := rack.HttpConnection(":3000")
	    rack.Up.Add(WorldWare{})
	    rack.Up.Add(HelloWare{})
	    rack.Run(conn, rack.Up)
	}
	
Then open up http://localhost:3000 to see "Hello World"

The following code will do the same thing a different way:

	package main

	import (
		"fmt"
		"github.com/HairyMezican/The-Rack/rack"
		"net/http"
	)

	func HelloWare(r *http.Request, vars rack.Vars, next rack.Next) (int, http.Header, []byte) {
		status, header, message := next()
		w := rack.CreateResponse(status, header, []byte{})
		fmt.Fprint(w, "Hello ", string(message))
		return w.Results()
	}

	func WorldWare(r *http.Request, vars rack.Vars, next rack.Next) (int, http.Header, []byte) {
		w := rack.BlankResponse()
		fmt.Fprint(w, "World")
		return w.Results()
	}

	func main() {
		conn := rack.HttpConnection(":3000")
		rack.Up.Add(rack.Func(HelloWare))
		rack.Up.Add(rack.Func(WorldWare))
		rack.Run(conn, rack.Up)
	}
	
and so will the following code:

	package main

	import (
		"github.com/HairyMezican/The-Rack/rack"
		"net/http"
	)

	func HelloWorldWare(r *http.Request, vars rack.Vars, next rack.Next) (int, http.Header, []byte) {
		return 200, make(http.Header), []byte("Hello World")
	}

	func main() {
		conn := rack.HttpConnection(":3000")
		rack.Run(conn, rack.Func(HelloWorldWare))
	}