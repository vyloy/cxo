package gnet

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/skycoin/cxo/node/log"
)

const TM time.Duration = 50 * time.Millisecond

// helper variables
var (
	tlsc = &tls.Config{InsecureSkipVerify: true}
)

func newConfig() (c Config) {
	c = NewConfig()
	if testing.Verbose() {
		c.Logger = log.NewLogger("[test] ", true)
	} else {
		c.Logger = log.NewLogger("[test] ", false)
		c.Logger.SetOutput(ioutil.Discard)
	}
	return
}

// helper functions
func dial(t *testing.T, address string) (c net.Conn) {
	var err error
	if c, err = net.DialTimeout("tcp", address, TM); err != nil {
		t.Fatal(err)
	}
	return
}

func readChan(t *testing.T, c chan struct{}, fatal string) {
	select {
	case <-c:
	case <-time.After(TM):
		t.Fatal(fatal)
	}
}

func TestNewPool(t *testing.T) {
	t.Run("invalid config", func(t *testing.T) {
		if _, err := NewPool(Config{MaxConnections: -1}); err == nil {
			t.Error("missing error")
		}
	})
	t.Run("logger", func(t *testing.T) {
		p, err := NewPool(NewConfig())
		if err != nil {
			t.Fatal(err)
		}
		if p.Logger == nil {
			t.Error("logger doen't created")
		}
		c := NewConfig()
		c.Logger = log.NewLogger("[asdf]", false)
		if p, err = NewPool(c); err != nil {
			t.Fatal(err)
		}
		if p.Logger != c.Logger {
			t.Error("another logger used")
		}
	})
}

func TestPool_Listen(t *testing.T) {
	t.Run("connect", func(t *testing.T) {
		connect := make(chan struct{})
		conf := newConfig()
		conf.ConnectionHandler = func(*Conn) { connect <- struct{}{} }
		p, err := NewPool(conf)
		if err != nil {
			t.Fatal(err)
		}
		defer p.Close()
		if err := p.Listen(""); err != nil {
			t.Error(err)
			return
		}
		if p.l == nil {
			t.Error("missing listener in Pool struct")
			return
		}
		c, err := net.DialTimeout("tcp", p.l.Addr().String(), TM)
		if err != nil {
			t.Error("error dialtig to the Pool")
			return
		}
		defer c.Close()
		readChan(t, connect, "slow connectiing")
	})
	t.Run("limit", func(t *testing.T) {
		//
	})
}

func TestPool_Address(t *testing.T) {
	//
}

func TestPool_Connections(t *testing.T) {
	//
}

func TestPool_Connection(t *testing.T) {
	//
}

func TestPool_Dial(t *testing.T) {
	//
}

func TestPool_Close(t *testing.T) {
	//
}