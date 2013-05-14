/*
	the HTTP bindings for the rack

	Whereas by default, the rack I've created is platform agnostic, httper helps add the http bindings.
	Rack middleware just defines a sequence of actions, http helps translate that into serving http requests.
*/
package httper

import (
	"bufio"
	"errors"
	"github.com/ScruffyProdigy/TheRack/rack"
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

type handler struct {
	m rack.Middleware
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	vars := rack.NewVars()
	vars[requestIndex] = r
	vars[originalRWIndex] = w
	h.m.Run(vars, func() {
		//none of the middleware handled the request
		//status is now "Not Found"
		if _, ok := vars[statusIndex]; !ok {
			vars[statusIndex] = http.StatusNotFound
		}

		if _, ok := vars[headerIndex]; !ok {
			vars[headerIndex] = NewHeader()
		}

		if _, ok := vars[messageIndex]; !ok {
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
}

type httpConnection struct {
	address string
}

func (this *httpConnection) Go(m rack.Middleware) error {
	return http.ListenAndServe(this.address, handler{m})
}

//HttpConnection() provides a basic HTTP Connection; good for a basic Website
func HttpConnection(addr string) rack.Connection {
	conn := new(httpConnection)
	conn.address = addr
	return conn
}

type httpsConnection struct {
	address  string
	certFile string
	keyFile  string
}

func (this *httpsConnection) Go(m rack.Middleware) error {
	return http.ListenAndServeTLS(this.address, this.certFile, this.keyFile, handler{m})
}

//HttpsConnection needs a certFile and a keyFile, but provides a more secure HTTPS connection for encrypted communication
func HttpsConnection(addr, certFile, keyFile string) rack.Connection {
	conn := new(httpsConnection)
	conn.address = addr
	conn.certFile = certFile
	conn.keyFile = keyFile

	return conn
}

//NewHeader is useful if you need a new blank header
func NewHeader() http.Header {
	return make(http.Header)
}

//V is a courtesy type - cast vars to this in order to get several manipulation functions
type V map[string]interface{}

//GetRequest gets the HTTP request that was sent
func (vars V) GetRequest() *http.Request {
	result, ok := vars[requestIndex].(*http.Request)
	if !ok {
		return nil
	}
	return result
}

//SetRequest sets the HTTP request to be used
func (vars V) SetRequest(r *http.Request) {
	vars[requestIndex] = r
}

//GetMessage gets the message that will be sent back
func (vars V) GetMessage() []byte {
	result, ok := vars[messageIndex].([]byte)
	if !ok {
		return nil
	}
	return result
}

//ResetMessage erases the message that will be sent back, and returns the old message
func (vars V) ResetMessage() []byte {
	result := vars.GetMessage()
	vars[messageIndex] = []byte{}
	return result
}

//SetMessage sets the message that will be sent back
func (vars V) SetMessage(message []byte) {
	vars[messageIndex] = message
}

//SetMessageString sets the message that will be sent back (but takes a string instead of a []byte)
func (vars V) SetMessageString(message string) {
	vars.SetMessage([]byte(message))
}

//AppendMessage adds to the message that will be sent back
func (vars V) AppendMessage(message []byte) {
	old := vars.GetMessage()
	if old == nil {
		old = []byte{}
	}
	vars[messageIndex] = append(old, message...)
}

//AppendMessageString adds to the message that will be sent back (but takes a string)
func (vars V) AppendMessageString(message string) {
	vars.AppendMessage([]byte(message))
}

//GetStatus gets the status that will be returned
func (vars V) GetStatus() int {
	result, ok := vars[statusIndex].(int)
	if !ok {
		return 0
	}
	return result
}

//Status sets the status that will be returned
func (vars V) Status(status int) {
	vars[statusIndex] = status
}

//StatusOK sets the status that will be returned to 200 - OK
//Should be the default status code
func (vars V) StatusOK() {
	vars.Status(http.StatusOK)
}

//StatusRedirect sets the status that will be returned to 301 - Found
//Should be the default redirect status code
func (vars V) StatusRedirect() {
	vars.Status(http.StatusFound)
}

//StatusNotFound sets the status that will be returned to 404 - Not Found
//Should be the default redirect code when you don't have anything to return
func (vars V) StatusNotFound() {
	vars.Status(http.StatusNotFound)
}

//StatusError sets the status that will be returned to 500 - InternalServerError
//Should be the default error status code
func (vars V) StatusError() {
	vars.Status(http.StatusInternalServerError)
}

//GetHeaders gets the headers that will be sent back
func (vars V) GetHeaders() http.Header {
	h, ok := vars[headerIndex].(http.Header)
	if !ok {
		h = NewHeader()
		vars[headerIndex] = h
	}
	return h
}

//SetHeaders sets the headers that will be sent back
func (vars V) SetHeaders(h http.Header) {
	vars[headerIndex] = h
}

//AddHeader adds a header to the list of headers that will be sent back
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

// FilledResponse creates a FakeResponseWriter out of values you already have
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

//Header() mimics a ResponseWriter's Header() function by returning the header we have
func (this *FakeResponseWriter) Header() http.Header {
	return this.header
}

//Write() mimics a ResponseWriter's Write() function by appending the new data to the data we have
func (this *FakeResponseWriter) Write(message []byte) (bytes int, err error) {
	this.message = append(this.message, message...)
	bytes += len(message)
	return
}

//WriteHeader() mimics a ResponseWriter's WriteHeader() function by storing the status they are trying to write
func (this *FakeResponseWriter) WriteHeader(status int) {
	this.status = status
}

//Hijack() can't entirely mimic the ResponseWriter's Hijack(),
//but we keep the original ResponseWriter around, and just call that one instead;
//also store that we've been hijacked, so we can ignore the result at the end
//(because once we've been hijacked, we are no longer responsible for the output)
func (this *FakeResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := this.vars[originalRWIndex].(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("Hijack not implemented")
	}

	this.vars[hijackedIndex] = true
	return hj.Hijack()
}

//Save saves the result of the ResponseWriter back to the vars
func (this *FakeResponseWriter) Save() {
	this.vars[statusIndex] = this.status
	this.vars[headerIndex] = this.header
	this.vars[messageIndex] = this.message
}

//Results returns the results of the ResponseWriter back to the vars
func (this *FakeResponseWriter) Results() (int, http.Header, []byte) {
	return this.status, this.header, this.message
}
