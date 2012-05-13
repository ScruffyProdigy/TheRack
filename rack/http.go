package rack

import (
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

func (this httpXConnection) Go(m Middleware) error {
	return this.c.Link(func(w http.ResponseWriter, r *http.Request) {
		vars := NewVars()
		vars[requestIndex] = r
		vars[originalRWIndex] = w
		m.Run(vars, func() {
			//none of the middleware handled the request
			//status is now "Not Found"
			vars.SetIfEmpty(statusIndex, http.StatusNotFound)
			vars.SetIfEmpty(headerIndex, NewHeader())
			vars.SetIfEmpty(messageIndex, []byte{})
		})

		hijacked, ok := vars.Get(hijackedIndex).(bool)
		if ok && hijacked {
			return
		}

		header, ok := vars.Get(headerIndex).(http.Header)
		if !ok {
			//default header is empty
			header = NewHeader()
		}

		status, ok := vars.Get(statusIndex).(int)
		if !ok {
			//one of the middleware handled the request, but forgot to set the status
			//status is now "OK"
			status = http.StatusOK
		}

		message, ok := vars.Get(messageIndex).([]byte)
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
func HttpConnection(addr string) Connection {
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
func HttpsConnection(addr, certFile, keyFile string) Connection {
	conn := new(httpsConnection)
	conn.address = addr
	conn.certFile = certFile
	conn.keyFile = keyFile

	return &httpXConnection{conn}
}

// FakeResponseWriter allows a rack program to take advantage of existing routines that require a ResponseWriter
// since we've eschewed the ResponseWriter path in favor of rack, we won't typically have one laying around
// but we can take what we do have, and fake the interface
// and then get the data back out so we can continue on
type fake struct {
	status  int
	header  http.Header
	message []byte
	vars    Vars
}

type FakeResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	Save()
}

// BlankResponse creates an entirely new FakeResponseWriter with default (mostly blank) values
func BlankResponse(vars Vars) FakeResponseWriter {
	this := new(fake)
	this.vars = vars
	this.status = http.StatusOK
	this.header = NewHeader()
	this.message = make([]byte, 0)
	return this
}

// CreateResponse creates a FakeResponseWriter out of values you already have
func CreateResponse(vars Vars) FakeResponseWriter {
	this := new(fake)

	this.vars = vars

	var ok bool
	this.status, ok = vars[statusIndex].(int)
	if !ok {
		this.status = http.StatusOK
	}

	this.header, ok = vars[headerIndex].(http.Header)
	if !ok {
		this.header = NewHeader()
	}

	this.message, ok = vars[messageIndex].([]byte)
	if !ok {
		this.message = make([]byte, 0)
	}

	return this
}

func (this *fake) Header() http.Header {
	return this.header
}

func (this *fake) Write(message []byte) (bytes int, err error) {
	this.message = append(this.message, message...)
	bytes += len(message)
	return
}

func (this *fake) WriteHeader(status int) {
	this.status = status
}

func (this *fake) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	this.vars[hijackedIndex] = true
	hj, ok := this.vars[originalRWIndex].(http.Hijacker)
	if !ok {
		return nil, nil, errors.New("Hijack not implemented")
	}

	return hj.Hijack()
}

func (this *fake) Save() {
	this.vars[statusIndex] = this.status
	this.vars[headerIndex] = this.header
	this.vars[messageIndex] = this.message
}

func NewHeader() http.Header {
	return make(http.Header)
}

func GetRequest(vars Vars) *http.Request {
	result, ok := vars[requestIndex].(*http.Request)
	if !ok {
		return nil
	}
	return result
}

func GetMessage(vars Vars) []byte {
	result, ok := vars[messageIndex].([]byte)
	if !ok {
		return nil
	}
	return result
}

func ResetMessage(vars Vars) []byte {
	result := GetMessage(vars)
	vars[messageIndex] = []byte{}
	return result
}

func SetMessage(vars Vars, message []byte) {
	vars[messageIndex] = message
}

func SetMessageString(vars Vars, message string) {
	SetMessage(vars, []byte(message))
}

func AppendMessage(vars Vars, message []byte) {
	old := GetMessage(vars)
	if old == nil {
		old = []byte{}
	}
	vars[messageIndex] = append(old, message...)
}

func AppendMessageString(vars Vars, message string) {
	AppendMessage(vars, []byte(message))
}

func GetStatus(vars Vars) int {
	result, ok := vars[statusIndex].(int)
	if !ok {
		return 0
	}
	return result
}

func Status(vars Vars, status int) {
	vars[statusIndex] = status
}

func StatusOK(vars Vars) {
	Status(vars, http.StatusOK)
}

func StatusRedirect(vars Vars) {
	Status(vars, http.StatusFound)
}

func StatusNotFound(vars Vars) {
	Status(vars, http.StatusNotFound)
}

func StatusError(vars Vars) {
	Status(vars, http.StatusInternalServerError)
}

func AddHeader(vars Vars, k, v string) {
	h, ok := vars[headerIndex].(http.Header)
	if !ok {
		h = NewHeader()
		vars[headerIndex] = h
	}
	h.Add(k, v)
}
