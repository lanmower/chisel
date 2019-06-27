package chshare

import (
	"log"
	"net"
	"time"
	"fmt"
	"github.com/gorilla/websocket"
)

type wsConn struct {
	*websocket.Conn
	buff []byte
}

func NewWebSocketConn(websocketConn *websocket.Conn) net.Conn {
	c := wsConn{
		Conn: websocketConn,
	}
	fmt.Println("init keepalive");
	keepAlive(c, 60000);
	return &c
}

func keepAlive(c *websocket.Conn, timeout time.Duration) {
    lastResponse := time.Now()
    c.SetPongHandler(func(msg string) error {
       lastResponse = time.Now()
       return nil
   })

   go func() {
     for {
	fmt.Println("sending keepalive");
        err := c.WriteMessage(websocket.PingMessage, []byte("keepalive"))
        if err != nil {
            return 
        }   
        time.Sleep(timeout/2)
	fmt.Println("last keepalive", time.Now().Sub(lastResponse));
        if(time.Now().Sub(lastResponse) > timeout) {
            c.Close()
            return
        }
    }
  }()
}


//Read is not threadsafe though thats okay since there
//should never be more than one reader
func (c *wsConn) Read(dst []byte) (int, error) {
	ldst := len(dst)
	//use buffer or read new message
	var src []byte
	if l := len(c.buff); l > 0 {
		src = c.buff
		c.buff = nil
	} else {
		t, msg, err := c.Conn.ReadMessage()
		if err != nil {
			return 0, err
		} else if t != websocket.BinaryMessage {
			log.Printf("<WARNING> non-binary msg")
		}
		src = msg
	}
	//copy src->dest
	var n int
	if len(src) > ldst {
		//copy as much as possible of src into dst
		n = copy(dst, src[:ldst])
		//copy remainder into buffer
		r := src[ldst:]
		lr := len(r)
		c.buff = make([]byte, lr)
		copy(c.buff, r)
	} else {
		//copy all of src into dst
		n = copy(dst, src)
	}
	//return bytes copied
	return n, nil
}

func (c *wsConn) Write(b []byte) (int, error) {
	if err := c.Conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	n := len(b)
	return n, nil
}

func (c *wsConn) SetDeadline(t time.Time) error {
	if err := c.Conn.SetReadDeadline(t); err != nil {
		return err
	}
	return c.Conn.SetWriteDeadline(t)
}
