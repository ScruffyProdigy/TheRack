package rack

import (
	"net/http"
)

type fake struct {
	status  int
	header  http.Header
	message []byte
}

// FakeResponseWriter allows a rack program to take advantage of existing routines that require a ResponseWriter
// since we've eschewed the ResponseWriter path in favor of rack, we won't typically have one laying around
// but we can take what we do have, and fake the interface
// and then get the data back out so we can continue on
type FakeResponseWriter interface {
	http.ResponseWriter
	Results() (int, http.Header, []byte) //Results lets us get the values back out of the ResponseWriter so that we, or another piece of Middleware can adjust them
}

// BlankResponse creates an entirely new FakeResponseWriter with default (mostly blank) values
func BlankResponse() FakeResponseWriter {
	this := new(fake)
	this.status = http.StatusOK
	this.header = NewHeader()
	this.message = make([]byte, 0)
	return this
}

// CreateResponse creates a FakeResponseWriter out of values you already have
func CreateResponse(status int, header http.Header, message []byte) FakeResponseWriter {
	this := new(fake)
	this.status = status
	this.header = header
	this.message = message
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

func (this *fake) Results() (int, http.Header, []byte) {
	return this.status, this.header, this.message
}
