package rack

import "net/http"

//Connection provides a common interface for both HTTP and HTTPS Connections
type Connection interface {
	Go(func(http.ResponseWriter, *http.Request)) error //Once you have the connection, just call go with a function that can handle a Response and a Request
}

type httpConnection struct {
	address string
}

func (this *httpConnection) Go(f func(http.ResponseWriter, *http.Request)) error {
	http.HandleFunc("/", f)
	return http.ListenAndServe(this.address, nil)
}

//HttpConnection provides a basic HTTP Connection; good for a basic Website
func HttpConnection(addr string) Connection {
	conn := new(httpConnection)
	conn.address = addr
	return conn
}

type httpsConnection struct {
	address  string
	certFile string
	keyFile  string
}

func (this *httpsConnection) Go(f func(http.ResponseWriter, *http.Request)) error {
	http.HandleFunc("/", f)
	return http.ListenAndServeTLS(this.address, this.certFile, this.keyFile, nil)
}

//HttpsConnection needs a certFile and a keyFile, but provides a more secure Https connection for encrypted communication
func HttpsConnection(addr, certFile, keyFile string) Connection {
	conn := new(httpsConnection)
	conn.address = addr
	conn.certFile = certFile
	conn.keyFile = keyFile

	return conn
}
