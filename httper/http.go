package httper

import (
	"github.com/HairyMezican/TheRack/rack"
	"bufio"
	"errors"
	"net"
	"net/http"
)

const (
	requestIndex    = "http.Request"
	originalRWIndex = "http.Original"
	statusIndex     = "http.Status"
	headerIndex     = "http.Header"
	messageIndex    = "http.Message"
	hijackedIndex   = "http.Hijacked"
)

//The httpXConnection provides a common interface for both HTTP and HTTPS connections
type httpXConnection struct {
	c httpxConnection
}

type httpxConnection interface {
	Link(func(http.ResponseWriter, *http.Request)) error //Once you have the connection, just call go with a function that can handle a Response and a Request
}

func (this httpXConnection) Go(m rack.Middleware) error {
	return this.c.Link(func(w http.ResponseWriter, r *http.Request) {
		vars := rack.NewVars()
		vars[requestIndex] = r
		vars[originalRWIndex] = w
		m.Run(vars, func() {
			//none of the middleware handled the request
			//status is now "Not Found"
			if _,ok := vars[statusIndex];!ok {
				vars[statusIndex] = http.StatusNotFound
			}
			
			if _,ok := vars[headerIndex];!ok {
				vars[headerIndex] = NewHeader()
			}
			
			if _,ok := vars[messageIndex];!ok {
				vars[messageIndex] = []byte{}
			}
		})

		hijacked, ok := vars[hijackedIndex].(bool)
		if ok && hijacked {
			return
		}

		header, ok := vars[headerIndex].(http.Header)
		if !ok {
			//default header is empty
			header = NewHeader()
		}

		status, ok := vars[statusIndex].(int)
		if !ok {
			//one of the middleware handled the request, but forgot to set the status
			//status is now "OK"
			status = http.StatusOK
		}

		message, ok := vars[messageIndex].([]byte)
		if !ok {
			//default message is blank
			message = []byte{}
		}

		for k, _ := range header {
			w.Header().Set(k, header.Get(k))
		}
		w.WriteHeader(status)
		w.Write(message)
	})
}

type httpConnection struct {
	address string
}

func (this *httpConnection) Link(f func(http.ResponseWriter, *http.Request)) error {
	http.HandleFunc("/", f)
	return http.ListenAndServe(this.address, nil)
}

//HttpConnection provides a basic HTTP Connection; good for a basic Website
func HttpConnection(addr string) rack.Connection {
	conn := new(httpConnection)
	conn.address = addr
	return &httpXConnection{conn}
}

type httpsConnection struct {
	address  string
	certFile string
	keyFile  string
}

func (this *httpsConnection) Link(f func(http.ResponseWriter, *http.Request)) error {
	http.HandleFunc("/", f)
	return http.ListenAndServeTLS(this.address, this.certFile, this.keyFile, nil)
}

//HttpsConnection needs a certFile and a keyFile, but provides a more secure Https connection for encrypted communication
func HttpsConnection(addr, certFile, keyFile string) rack.Connection {
	conn := new(httpsConnection)
	conn.address = addr
	conn.certFile = certFile
	conn.keyFile = keyFile

	return &httpXConnection{conn}
}

func NewHeader() http.Header {
	return make(http.Header)
}

type V map[string]interface{}

func (vars V) GetRequest() *http.Request {
	result, ok := vars[requestIndex].(*http.Request)
	if !ok {
		return nil
	}
	return result
}

func (vars V) SetRequest(r *http.Request) {
	vars[requestIndex] = r
} 

func (vars V) GetMessage() []byte {
	result, ok := vars[messageIndex].([]byte)
	if !ok {
		return nil
	}
	return result
}

func (vars V) ResetMessage() []byte {
	result := vars.GetMessage()
	vars[messageIndex] = []byte{}
	return result
}

func (vars V) SetMessage(message []byte) {
	vars[messageIndex] = message
}

func (vars V) SetMessageString(message string) {
	vars.SetMessage([]byte(message))
}

func (vars V) AppendMessage(message []byte) {
	old := vars.GetMessage()
	if old == nil {
		old = []byte{}
	}
	vars[messageIndex] = append(old, message...)
}

func (vars V) AppendMessageString(message string) {
	vars.AppendMessage([]byte(message))
}

func (vars V) GetStatus() int {
	result, ok := vars[statusIndex].(int)
	if !ok {
		return 0
	}
	return result
}

func (vars V) Status(status int) {
	vars[statusIndex] = status
}

func (vars V) StatusOK() {
	vars.Status(http.StatusOK)
}

func (vars V) StatusRedirect() {
	vars.Status(http.StatusFound)
}

func (vars V) StatusNotFound() {
	vars.Status(http.StatusNotFound)
}

func (vars V) StatusError() {
	vars.Status(http.StatusInternalServerError)
}

func (vars V) GetHeaders() http.Header {
	h,ok := vars[headerIndex].(http.Header)
	if !ok {
		h = NewHeader()
		vars[headerIndex] = h
	}
	return h
}

func (vars V) SetHeaders(h http.Header) {
	vars[headerIndex] = h
}

func (vars V) AddHeader(k, v string) {
	h, ok := vars[headerIndex].(http.Header)
	if !ok {
		h = NewHeader()
		vars[headerIndex] = h
	}
	h.Add(k, v)
}


// FakeResponseWriter allows a rack program to take advantage of existing routines that require a ResponseWriter
// since we've eschewed the ResponseWriter path in favor of rack, we won't typically have one laying around
// but we can take what we do have, and fake the interface
// and then get the data back out so we can continue on

type FakeResponseWriter struct {
	status  int
	header  http.Header
	message []byte
	vars    V
}


// BlankResponse creates an entirely new FakeResponseWriter with default (mostly blank) values
func (vars V) BlankResponse() *FakeResponseWriter {
	w := new(FakeResponseWriter)
	w.vars = vars
	w.status = http.StatusOK
	w.header = NewHeader()
	w.message = make([]byte, 0)
	return w
}

// CreateResponse creates a FakeResponseWriter out of values you already have
func (vars V) FilledResponse() *FakeResponseWriter {
	w := new(FakeResponseWriter)

	w.vars = vars

	var ok bool
	w.status, ok = vars[statusIndex].(int)
	if !ok {
		w.status = http.StatusOK
	}

	w.header, ok = vars[headerIndex].(http.Header)
	if !ok {
		w.header = NewHeader()
	}

	w.message, ok = vars[messageIndex].([]byte)
	if !ok {
		w.message = make([]byte, 0)
	}

	return w
}

func (this *FakeResponseWriter) Header() http.Header {
	return this.header
}

func (this *FakeResponseWriter) Write(message []byte) (bytes int, err error) {
	this.message = append(this.message, message...)
	bytes += len(message)
	return
}

func (this *FakeResponseWriter) WriteHeader(status int) {
	this.status = status
}

func (this *FakeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	this.vars[hijackedIndex] = true
	hj, ok := this.vars[originalRWIndex].(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("Hijack not implemented")
	}

	return hj.Hijack()
}

func (this *FakeResponseWriter) Save() {
	this.vars[statusIndex] = this.status
	this.vars[headerIndex] = this.header
	this.vars[messageIndex] = this.message
}

func (this *FakeResponseWriter) Results() (int,http.Header,[]byte) {
	return this.status,this.header,this.message
}

type Next func() (int,http.Header,[]byte)

type Middleware interface {
	Serve(*http.Request,V,Next) (int,http.Header,[]byte)
}

type Func func(*http.Request,V,Next) (int,http.Header,[]byte)

func (this Func) Serve(r *http.Request,vars V,next Next) (int,http.Header,[]byte) {
	return this(r,vars,next)
} 

func (this Func) Run(vars map[string] interface{},next func()) {
	Racker{this}.Run(vars,next)
}

type Racker struct{
	m Middleware
}

func (this Racker) Run(vars map[string] interface{},next func()) {
	v := V(vars)
	status,header,message := this.m.Serve(v.GetRequest(),v,func() (int, http.Header,[]byte) {
		next()
		return v.GetStatus(),v.GetHeaders(),v.GetMessage()
	})
	v.Status(status)
	v.SetHeaders(header)
	v.SetMessage(message)
}

type Deracker struct{
	m rack.Middleware
}

func (this Deracker) Serve(r *http.Request, vars V, next Next) (int,http.Header,[]byte) {
	vars.SetRequest(r)
	this.m.Run(vars,func() {
		status,header,message := next()
		vars.Status(status)
		vars.SetHeaders(header)
		vars.SetMessage(message)
	})
	return vars.GetStatus(),vars.GetHeaders(),vars.GetMessage()
}

type Rack []Middleware

func (this Rack) Serve(r *http.Request, vars V, next Next) (int,http.Header,[]byte) {
	index := -1
	var ourNext Next
	ourNext = func() (int,http.Header,[]byte) {
		index++
		defer func(){index--}()
		
		if index >= len(this) {
			return next()
		}
		return this[index].Serve(r, vars, ourNext)
	}
	return ourNext()
}

func (this Rack) Run(vars map[string]interface{}, next func()) {
	Racker{this}.Run(vars,next)
}

func NewRack() Rack {
	return make(Rack,0,2)
}