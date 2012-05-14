# Rack

A simple rack implementation in Go

Please see http://github.com/HairyMezican/Middleware for some usable premade middleware

run `go get github.com/HairyMezican/TheRack/rack` to install

## Examples

	package main

	import (
		"../TheRack/httper"
		"../TheRack/rack"
	)

	type HelloWare struct{}

	func (HelloWare) Run(vars map[string]interface{}, next func()) {
		(httper.V)(vars).SetMessageString("Hello " + vars["Object"].(string))
	}

	type WorldWare struct{}

	func (WorldWare) Run(vars map[string]interface{}, next func()) {
		vars["Object"] = "World"
		next()
	}

	func main() {
		rackup := rack.New()
		rackup.Add(WorldWare{})
		rackup.Add(HelloWare{})

		conn := httper.HttpConnection(":3000")
		conn.Go(rackup)
	}
	
Then open up http://localhost:3000 to see "Hello World"

The following code will do the same thing a different way:

	package main

	import (
		"../TheRack/httper"
		"../TheRack/rack"
		"fmt"
		"net/http"
	)

	func HelloWare(vars map[string]interface{}, next func()) {
		next()

		v := httper.V(vars)

		old := v.ResetMessage()
		v.SetMessageString("Hello ")
		v.AppendMessage(old)
	}

	func WorldWare(r *http.Request, vars httper.V, next httper.Next) (int, http.Header, []byte) {
		w := vars.BlankResponse()
		fmt.Fprint(w, "World")
		return w.Results()
	}

	func main() {
		rackup := rack.New()
		rackup.Add(rack.Func(HelloWare))
		rackup.Add(httper.Func(WorldWare))

		conn := httper.HttpConnection(":3000")
		conn.Go(rackup)
	}
	
and if you just want something short and simple, so will the following code:

	package main

	import (
		"../TheRack/httper"
		"../TheRack/rack"
	)

	var HelloWorldWare rack.Func = func(vars map[string]interface{}, next func()) {
		httper.V(vars).SetMessageString("Hello World")
	}

	func main() {
		conn := httper.HttpConnection(":3000")
		conn.Go(HelloWorldWare)
	}
	