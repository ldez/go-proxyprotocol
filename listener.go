package proxyprotocol

import (
	"bufio"
	"net"
	"sync"
	"time"
)

const bufferSize = 1400

// LoggerFn type of logger function
type LoggerFn func(string, ...interface{})

// SourceChecker check trusted address
type SourceChecker func(net.Addr) (bool, error)

// NewListener create new proxyprocol.Listener from any net.Listener.
func NewListener(listener net.Listener) *Listener {
	return &Listener{
		Listener: listener,
		HeaderParsers: []HeaderParser{
			ParseTextHeader,
			ParseBinaryHeader,
		},
	}
}

// Listener implement net.Listener
type Listener struct {
	Listener      net.Listener
	LoggerFn      LoggerFn
	HeaderParsers []HeaderParser
	SourceCheck   SourceChecker
}

func (listener *Listener) log(str string, args ...interface{}) {
	if nil != listener.LoggerFn {
		listener.LoggerFn(str, args...)
	}
}

// WithLogger copy Listener and set LoggerFn
func (listener *Listener) WithLogger(loggerFn LoggerFn) *Listener {
	newListener := *listener
	newListener.LoggerFn = loggerFn
	return &newListener
}

// WithHeaderParsers copy Listener and set HeaderParser.
// Can be used to disable or reorder HeaderParser's.
func (listener *Listener) WithHeaderParsers(headerParser ...HeaderParser) *Listener {
	newListener := *listener
	newListener.HeaderParsers = headerParser
	return &newListener
}

// WithSourceChecker copy Listener and set SourceChecker
func (listener *Listener) WithSourceChecker(sourceChecker SourceChecker) *Listener {
	newListener := *listener
	newListener.SourceCheck = sourceChecker
	return &newListener
}

// Accept implement net.Listener.Accept().
// If request have proxyprotocol header, then wrap connection into proxyprotocol.Conn.
// Otherwise return raw net.Conn.
func (listener *Listener) Accept() (net.Conn, error) {
	rawConn, err := listener.Listener.Accept()
	if nil != err {
		return nil, err
	}

	if listener.SourceCheck != nil {
		trusted, err := listener.SourceCheck(rawConn.RemoteAddr())
		if nil != err {
			listener.log("Source check error: %s", err)
			return nil, err
		}
		if !trusted {
			return rawConn, nil
		}
	}

	return NewConn(rawConn, listener.log, listener.HeaderParsers), nil
}

// Close is proxy to listener.Close()
func (listener *Listener) Close() error {
	return listener.Listener.Close()
}

// Addr is proxy to listener.Addr()
func (listener Listener) Addr() net.Addr {
	return listener.Listener.Addr()
}

// Conn is wrapper on net.Conn with overrided RemoteAddr()
type Conn struct {
	conn          net.Conn
	logf          LoggerFn
	readBuf       *bufio.Reader
	header        *Header
	headerParsers []HeaderParser
	once          sync.Once
}

// NewConn create wrapper on net.Conn.
// If proxyprtocol header is local, when header should be nil.
func NewConn(
	conn net.Conn,
	logf LoggerFn,
	headerParsers []HeaderParser,
) net.Conn {
	readBuf := bufio.NewReaderSize(conn, bufferSize)

	return &Conn{
		conn:          conn,
		readBuf:       readBuf,
		logf:          logf,
		headerParsers: headerParsers,
	}
}

func (conn *Conn) parserHeader() {
	for _, headerParser := range conn.headerParsers {
		header, err := headerParser(conn.readBuf, conn.logf)
		switch err {
		case nil:
			conn.logf("Use header remote addr")
			conn.header = header
			return
		case ErrInvalidSignature:
			continue
		default:
			conn.logf("Parse header error: %s", err)
			return
		}
	}
	conn.logf("Use raw remote addr")
}

// Read proxy to conn.Read
func (conn *Conn) Read(buf []byte) (int, error) {
	conn.once.Do(conn.parserHeader)
	return conn.readBuf.Read(buf)
}

// Write proxy to conn.Write
func (conn *Conn) Write(buf []byte) (int, error) {
	return conn.conn.Write(buf)
}

// Close proxy to conn.Close
func (conn *Conn) Close() error {
	return conn.conn.Close()
}

// LocalAddr proxy to conn.LocalAddr
func (conn *Conn) LocalAddr() net.Addr {
	return conn.conn.LocalAddr()
}

// RemoteAddr return addr of remote client.
// If proxyprtocol not local, then return src from header.
func (conn *Conn) RemoteAddr() net.Addr {
	conn.once.Do(conn.parserHeader)
	if nil != conn.header {
		return conn.header.SrcAddr
	}
	return conn.conn.RemoteAddr()
}

// SetDeadline proxy to conn.SetDeadline
func (conn *Conn) SetDeadline(t time.Time) error {
	return conn.conn.SetDeadline(t)
}

// SetReadDeadline proxy to conn.SetReadDeadline
func (conn *Conn) SetReadDeadline(t time.Time) error {
	return conn.conn.SetReadDeadline(t)
}

// SetWriteDeadline  proxy to conn.SetWriteDeadline
func (conn *Conn) SetWriteDeadline(t time.Time) error {
	return conn.conn.SetWriteDeadline(t)
}
