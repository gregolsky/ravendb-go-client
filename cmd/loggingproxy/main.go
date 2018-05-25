package main

// based on https://raw.githubusercontent.com/elazarl/goproxy/master/examples/goproxy-httpdump/httpdump.go

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"sync"

	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/transport"
)

var (
	tr    = transport.Transport{Proxy: transport.ProxyFromEnvironment}
	muLog sync.Mutex
)

func panicIf(cond bool, format string, args ...interface{}) {
	if cond {
		s := fmt.Sprintf(format, args...)
		panic(s)
	}
}

// TeeReadCloser extends io.TeeReader by allowing reader and writer to be
// closed.
type TeeReadCloser struct {
	r io.Reader
	w io.WriteCloser
	c io.Closer
}

func NewTeeReadCloser(r io.ReadCloser, w io.WriteCloser) io.ReadCloser {
	panicIf(r == nil, "r == nil")
	panicIf(w == nil, "w == nil")
	return &TeeReadCloser{io.TeeReader(r, w), w, r}
}

func (t *TeeReadCloser) Read(b []byte) (int, error) {
	panicIf(t == nil, "t == nil")
	panicIf(t.r == nil, "t.r == nil, t: %#v", t)
	return t.r.Read(b)
}

// Close attempts to close the reader and write. It returns an error if both
// failed to Close.
func (t *TeeReadCloser) Close() error {
	err1 := t.c.Close()
	err2 := t.w.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

// stoppableListener serves stoppableConn and tracks their lifetime to notify
// when it is safe to terminate the application.
type stoppableListener struct {
	net.Listener
	sync.WaitGroup
}

type stoppableConn struct {
	net.Conn
	wg *sync.WaitGroup
}

func newStoppableListener(l net.Listener) *stoppableListener {
	return &stoppableListener{l, sync.WaitGroup{}}
}

func (sl *stoppableListener) Accept() (net.Conn, error) {
	c, err := sl.Listener.Accept()
	if err != nil {
		return c, err
	}
	sl.Add(1)
	return &stoppableConn{c, &sl.WaitGroup}, nil
}

func (sc *stoppableConn) Close() error {
	sc.wg.Done()
	return sc.Conn.Close()
}

type BufferCloser struct {
	*bytes.Buffer
}

func NewBufferCloser(buf *bytes.Buffer) *BufferCloser {
	if buf == nil {
		buf = &bytes.Buffer{}
	}
	return &BufferCloser{
		Buffer: buf,
	}
}

func (b *BufferCloser) Close() error {
	// nothing to do
	return nil
}

// SessionData has info about
type SessionData struct {
	reqBody  *BufferCloser
	respBody *BufferCloser
}

func NewSessionData() *SessionData {
	return &SessionData{}
}

func handleOnRequest(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	panicIf(req == nil, "req == nil")
	sd := NewSessionData()
	ctx.UserData = sd

	if req.Body != nil {
		sd.reqBody = NewBufferCloser(nil)
		req.Body = NewTeeReadCloser(req.Body, sd.reqBody)
	}
	return req, nil
}

// if d is a valid json, pretty-print it
func prettyPrintMaybeJSON(d []byte) []byte {
	var m map[string]interface{}
	err := json.Unmarshal(d, &m)
	if err != nil {
		return d
	}
	d2, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return d
	}
	return d2
}

func lg(d []byte) {
	muLog.Lock()
	os.Stdout.Write(d)
	muLog.Unlock()
}

func lgReq(ctx *goproxy.ProxyCtx, reqBody []byte, respBody []byte) {
	reqBody = prettyPrintMaybeJSON(reqBody)
	respBody = prettyPrintMaybeJSON(respBody)

	var buf bytes.Buffer
	s := fmt.Sprintf("=========== %d:\n", ctx.Session)
	buf.WriteString(s)
	d, err := httputil.DumpRequest(ctx.Req, false)
	if err == nil {
		buf.Write(d)
	}
	buf.Write(reqBody)

	s = "\n--------\n"
	buf.WriteString(s)
	if ctx.Resp != nil {
		d, err = httputil.DumpResponse(ctx.Resp, false)
		if err == nil {
			buf.Write(d)
		}
		buf.Write(respBody)
		buf.WriteString("\n")
	}

	lg(buf.Bytes())
}

func slurpResponseBody(resp *http.Response) []byte {
	if resp == nil {
		return nil
	}
	d, err := ioutil.ReadAll(resp.Body)
	panicIf(err != nil, "err: %v", err)
	resp.Body = NewBufferCloser(bytes.NewBuffer(d))
	return d
}

func handleOnResponse(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	panicIf(resp != ctx.Resp, "resp != ctx.Resp")

	sd := ctx.UserData.(*SessionData)
	reqBody := sd.reqBody.Bytes()
	respBody := slurpResponseBody(resp)
	lgReq(ctx, reqBody, respBody)

	return resp
}

func main() {
	verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
	addr := flag.String("l", ":8888", "on which address should the proxy listen")
	flag.Parse()
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *verbose
	if err := os.MkdirAll("db", 0755); err != nil {
		log.Fatal("Can't create dir", err)
	}

	proxy.OnRequest().DoFunc(handleOnRequest)
	proxy.OnResponse().DoFunc(handleOnResponse)
	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal("listen:", err)
	}
	sl := newStoppableListener(l)
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	go func() {
		<-ch
		log.Println("Got SIGINT exiting")
		sl.Add(1)
		sl.Close()
		//logger.Close()
		sl.Done()
	}()
	log.Println("Starting Proxy")
	http.Serve(sl, proxy)
	sl.Wait()
	log.Println("All connections closed - exit")
}
