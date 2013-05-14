package rack_test

import (
	"./httper"
	"./rack"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetFrom(loc string) {
	resp, err := http.Get(loc)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	fmt.Println(string(body))
}

/*Example 1*/
type HelloWare1 struct{}

func (HelloWare1) Run(vars map[string]interface{}, next func()) {
	(httper.V)(vars).SetMessageString("Hello " + vars["Object"].(string))
}

type WorldWare1 struct{}

func (WorldWare1) Run(vars map[string]interface{}, next func()) {
	vars["Object"] = "World"
	next()
}

func Example_1() {
	rackup := rack.New()
	rackup.Add(WorldWare1{})
	rackup.Add(HelloWare1{})

	conn := httper.HttpConnection(":3001")

	go func() {
		conn.Go(rackup)
	}()

	GetFrom("http://localhost:3001")
	//output: Hello World
}

/*Example 2*/
func HelloWare2(vars map[string]interface{}, next func()) {
	next()

	v := httper.V(vars)

	old := v.ResetMessage()
	v.SetMessageString("Hello ")
	v.AppendMessage(old)
}

func WorldWare2(vars map[string]interface{}, next func()) {
	w := httper.V(vars).BlankResponse()
	fmt.Fprint(w, "World")
	w.Save()
}

func Example_2() {
	rackup := rack.New()
	rackup.Add(rack.Func(HelloWare2))
	rackup.Add(rack.Func(WorldWare2))
	conn := httper.HttpConnection(":3002")

	go func() {
		conn.Go(rackup)
	}()

	GetFrom("http://localhost:3002")
	//output: Hello World
}

/*Example 3*/

var HelloWorldWare3 rack.Func = func(vars map[string]interface{}, next func()) {
	httper.V(vars).SetMessageString("Hello World")
}

func Example_3() {

	conn := httper.HttpConnection(":3003")

	go func() {
		conn.Go(HelloWorldWare3)
	}()

	GetFrom("http://localhost:3003")
	//output: Hello World
}
