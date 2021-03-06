package proxyprotocol_test

import (
	"bufio"
	"errors"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/c0va23/go-proxyprotocol"
	gomock "github.com/golang/mock/gomock"
)

func TestConn_Read(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	rawAddr := &net.TCPAddr{
		IP:   net.IPv4(10, 0, 0, 1),
		Port: 12345,
	}

	rawConn.EXPECT().RemoteAddr().Return(rawAddr).AnyTimes()

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	readBuf := bufio.NewReaderSize(rawConn, 1400)

	buf := make([]byte, readBuf.Size())

	t.Run("when header parser return err", func(t *testing.T) {
		parseErr := errors.New("Parse error")
		headerParser.EXPECT().Parse(readBuf).Return(nil, parseErr)
		logger.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

		trustedAddr := true
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

		n, err := conn.Read(buf)

		if err != parseErr {
			t.Errorf("Unexpected error %s", err)
		}

		if n != 0 {
			t.Errorf("Unexpected read size %d", n)
		}

		if srcAddr := conn.RemoteAddr(); !reflect.DeepEqual(srcAddr, rawAddr) {
			t.Errorf("Unexpected remote addr %s", srcAddr)
		}

		t.Run("when call conn.Read second time", func(t *testing.T) {
			_, err := conn.Read(buf)

			if err != parseErr {
				t.Errorf("Unexpected error %s", err)
			}
		})
	})

	t.Run("when header parser return Header", func(t *testing.T) {
		header := &proxyprotocol.Header{
			SrcAddr: &net.TCPAddr{
				IP: net.IPv4(1, 2, 3, 4),
			},
		}

		headerParser.EXPECT().Parse(readBuf).Return(header, nil).AnyTimes()

		t.Run("when rawConn.Read return err", func(t *testing.T) {
			readErr := errors.New("Read error")
			rawConn.EXPECT().Read(buf).Return(0, readErr)

			trustedAddr := true
			conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

			n, err := conn.Read(buf)
			if err != readErr {
				t.Errorf("Unexpected error %s", err)
			}

			if n != 0 {
				t.Errorf("Unexpected read size %d", n)
			}

			if srcAddr := conn.RemoteAddr(); !reflect.DeepEqual(srcAddr, header.SrcAddr) {
				t.Errorf("Unexpected remote addr %s", srcAddr)
			}
		})

		t.Run("when rawConn.Read return read size", func(t *testing.T) {
			readSize := 512
			rawConn.EXPECT().Read(buf).Return(readSize, nil)

			trustedAddr := true
			conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

			n, err := conn.Read(buf)

			if err != nil {
				t.Errorf("Unexpected error %s", err)
			}

			if n != readSize {
				t.Errorf("Unexpected read size %d", readSize)
			}

			if srcAddr := conn.RemoteAddr(); !reflect.DeepEqual(srcAddr, header.SrcAddr) {
				t.Errorf("Unexpected remote addr %s", srcAddr)
			}
		})
	})
}

func TestConnRemoteAddr(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	rawAddr := &net.TCPAddr{
		IP:   net.IPv4(10, 0, 0, 1),
		Port: 12345,
	}

	rawConn.EXPECT().RemoteAddr().Return(rawAddr).AnyTimes()

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	readBuf := bufio.NewReaderSize(rawConn, 1400)

	t.Run("when header parser return error", func(t *testing.T) {
		trustedAddr := true
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

		parseErr := errors.New("Parse error")

		headerParser.EXPECT().Parse(readBuf).Return(nil, parseErr)
		logger.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

		remoteAddr := conn.RemoteAddr()

		if !reflect.DeepEqual(remoteAddr, rawAddr) {
			t.Errorf("Unexpected remote addr %s", remoteAddr)
		}
	})

	t.Run("when header parser return header", func(t *testing.T) {
		header := &proxyprotocol.Header{
			SrcAddr: &net.TCPAddr{
				IP: net.IPv4(1, 2, 3, 4),
			},
		}

		headerParser.EXPECT().Parse(readBuf).Return(header, nil).AnyTimes()

		t.Run("when trusted addr", func(t *testing.T) {
			trustedAddr := true
			conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

			remoteAddr := conn.RemoteAddr()

			if !reflect.DeepEqual(remoteAddr, header.SrcAddr) {
				t.Errorf("Unexpected remote adder %s", remoteAddr)
			}

			t.Run("when second call .RemoteAddr()", func(t *testing.T) {
				remoteAddr := conn.RemoteAddr()

				if !reflect.DeepEqual(remoteAddr, header.SrcAddr) {
					t.Errorf("Unexpected remote adder %s", remoteAddr)
				}
			})
		})

		t.Run("when not trusted addr", func(t *testing.T) {
			trustedAddr := false
			conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

			remoteAddr := conn.RemoteAddr()

			if !reflect.DeepEqual(remoteAddr, rawAddr) {
				t.Errorf("Unexpected remote adder %s", remoteAddr)
			}
		})
	})
}

func TestConn_Close(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	trustedAddr := true
	conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

	t.Run("when rawConn.Close() return error", func(t *testing.T) {
		closeErr := errors.New("Close error")
		rawConn.EXPECT().Close().Return(closeErr)

		err := conn.Close()

		if err != closeErr {
			t.Errorf("Unexpected error %s", err)
		}
	})

	t.Run("when rawConn.Close() return nil", func(t *testing.T) {
		rawConn.EXPECT().Close().Return(nil)

		err := conn.Close()

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	})
}

func TestConn_LocalAddr(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	logger.EXPECT().Printf(gomock.Any(), gomock.Any()).AnyTimes()

	rawAddr := &net.TCPAddr{
		IP:   net.IPv4(10, 0, 0, 1),
		Port: 12345,
	}
	rawConn.EXPECT().LocalAddr().Return(rawAddr).AnyTimes()

	readBuf := bufio.NewReaderSize(rawConn, 1400)

	t.Run("when header parser return nil header", func(t *testing.T) {
		headerParser.EXPECT().Parse(readBuf).Return(nil, nil)

		trustedAddr := true
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

		localAddr := conn.LocalAddr()
		if localAddr != rawAddr {
			t.Errorf("Unexpected local addr %s", localAddr)
		}
	})

	t.Run("when header parser return not nil header", func(t *testing.T) {
		trustedAddr := true
		conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

		header := &proxyprotocol.Header{
			DstAddr: &net.TCPAddr{
				IP:   net.IPv4(1, 2, 3, 4),
				Port: 12345,
			},
		}
		headerParser.EXPECT().Parse(readBuf).Return(header, nil).AnyTimes()

		localAddr := conn.LocalAddr()
		if localAddr != header.DstAddr {
			t.Errorf("Unexpected local addr %s", localAddr)
		}

		t.Run("when second call LocalAddr()", func(t *testing.T) {
			t.Run("when trusted addr", func(t *testing.T) {
				trustedAddr := true
				conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

				localAddr := conn.LocalAddr()
				if localAddr != header.DstAddr {
					t.Errorf("Unexpected local addr %s", localAddr)
				}
			})

			t.Run("when trusted addr", func(t *testing.T) {
				trustedAddr := false
				conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

				localAddr := conn.LocalAddr()
				if localAddr != rawAddr {
					t.Errorf("Unexpected local addr %s", localAddr)
				}
			})
		})
	})
}

func TestConn_Write(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	trustedAddr := true
	conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

	buf := []byte{1, 2, 3, 4, 5}

	t.Run("when rawConn.Write return error", func(t *testing.T) {
		writeErr := errors.New("Write error")
		rawConn.EXPECT().Write(buf).Return(0, writeErr)

		n, err := conn.Write(buf)
		if err != writeErr {
			t.Errorf("Unexpected error %s", err)
		}

		if n != 0 {
			t.Errorf("Unexpected write size")
		}
	})

	t.Run("when rawConn.Write return write size", func(t *testing.T) {
		writeSize := len(buf)
		rawConn.EXPECT().Write(buf).Return(writeSize, nil)

		n, err := conn.Write(buf)

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}

		if n != writeSize {
			t.Errorf("Unexpected write size")
		}
	})
}

func TestConn_SetDeadline(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	trustedAddr := true
	conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

	deadLine := time.Now().Add(time.Second * 15)

	t.Run("when rawConn.SetDeadline return error", func(t *testing.T) {
		deadLineErr := errors.New("DeadLine error")
		rawConn.EXPECT().SetDeadline(deadLine).Return(deadLineErr)

		err := conn.SetDeadline(deadLine)

		if err != deadLineErr {
			t.Errorf("Unexpected error %s", err)
		}
	})

	t.Run("when rawConn.SetDeadline return nil", func(t *testing.T) {
		rawConn.EXPECT().SetDeadline(deadLine).Return(nil)

		err := conn.SetDeadline(deadLine)

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	})
}

func TestConn_SetReadDeadline(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	trustedAddr := true
	conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

	deadLine := time.Now().Add(time.Second * 15)

	t.Run("when rawConn.SetReadDeadline return error", func(t *testing.T) {
		deadLineErr := errors.New("ReadDeadLine error")
		rawConn.EXPECT().SetReadDeadline(deadLine).Return(deadLineErr)

		err := conn.SetReadDeadline(deadLine)

		if err != deadLineErr {
			t.Errorf("Unexpected error %s", err)
		}
	})

	t.Run("when rawConn.SetReadDeadline return nil", func(t *testing.T) {
		rawConn.EXPECT().SetReadDeadline(deadLine).Return(nil)

		err := conn.SetReadDeadline(deadLine)

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	})
}

func TestConn_SetWriteDeadline(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	rawConn := NewMockConn(mockCtrl)

	headerParser := NewMockHeaderParser(mockCtrl)
	logger := NewMockLogger(mockCtrl)

	trustedAddr := true
	conn := proxyprotocol.NewConn(rawConn, logger, headerParser, trustedAddr)

	deadLine := time.Now().Add(time.Second * 15)

	t.Run("when rawConn.SetWriteDeadline return error", func(t *testing.T) {
		deadLineErr := errors.New("WriteDeadLine error")
		rawConn.EXPECT().SetWriteDeadline(deadLine).Return(deadLineErr)

		err := conn.SetWriteDeadline(deadLine)

		if err != deadLineErr {
			t.Errorf("Unexpected error %s", err)
		}
	})

	t.Run("when rawConn.SetWriteDeadline return nil", func(t *testing.T) {
		rawConn.EXPECT().SetWriteDeadline(deadLine).Return(nil)

		err := conn.SetWriteDeadline(deadLine)

		if err != nil {
			t.Errorf("Unexpected error %s", err)
		}
	})
}
