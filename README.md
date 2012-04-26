# Rack

A simple rack implementation in Go

run `go get github.com/HairyMezican/The-Rack/rack` to install

## Example

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