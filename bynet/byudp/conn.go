package byudp

import (
	"bylib/bylog"
	"context"
	"net"
	"sync"
	"time"
)

const (
	// MessageTypeBytes is the length of type header.
	MessageTypeBytes = 4
	// MessageLenBytes is the length of length header.
	MessageLenBytes = 4
	// MessageMaxBytes is the maximum bytes allowed for application data.
	MessageMaxBytes = 1 << 23 // 8M
)
// MessageHandler is a combination of message and its handler function.
type MessageHandler struct {
	message Message
	handler HandlerFunc
}

// WriteCloser is the interface that groups Write and Close methods.
type WriteCloser interface {
	Write(Message) error
	Close()
}
// ClientConn represents a client connection to a TCP server.
type ClientConn struct {
	laddr	  string
	addr      string
	opts      options
	netid     int64
	rawConn   net.Conn
	once      *sync.Once
	wg        *sync.WaitGroup
	sendCh    chan []byte
	handlerCh chan MessageHandler
	timing    *TimingWheel
	mu        sync.Mutex // guards following
	name      string
	heart     int64
	pending   []int64
	ctx       context.Context
	cancel    context.CancelFunc
}


// NewClientConn returns a new client connection which has not started to
// serve requests yet.
func NewClientConn(netid int64, c net.Conn, opt ...ServerOption) *ClientConn {
	var opts options
	for _, o := range opt {
		o(&opts)
	}

	//必须传递编解码器.
	if opts.codec == nil {
		opts.codec = TypeLengthValueCodec{}
	}
	if opts.bufferSize <= 0 {
		opts.bufferSize = BufferSize256
	}
	return newClientConnWithOptions(netid, c, opts)
}

func newClientConnWithOptions(netid int64, c net.Conn, opts options) *ClientConn {
	cc := &ClientConn{
		laddr:	   c.LocalAddr().String(),
		addr:      c.RemoteAddr().String(),
		opts:      opts,
		netid:     netid,
		rawConn:   c,
		once:      &sync.Once{},
		wg:        &sync.WaitGroup{},
		sendCh:    make(chan []byte, opts.bufferSize),
		handlerCh: make(chan MessageHandler, opts.bufferSize),
		heart:     time.Now().UnixNano(),
	}
	bylog.Debug("local addr=%s remote addr=%s",cc.laddr,cc.addr)
	cc.ctx, cc.cancel = context.WithCancel(context.Background())
	cc.timing = NewTimingWheel(cc.ctx)
	cc.name = c.RemoteAddr().String()
	cc.pending = []int64{}
	return cc
}

// NetID returns the net ID of client connection.
func (cc *ClientConn) NetID() int64 {
	return cc.netid
}

// SetName sets the name of client connection.
func (cc *ClientConn) SetName(name string) {
	cc.mu.Lock()
	cc.name = name
	cc.mu.Unlock()
}

// Name gets the name of client connection.
func (cc *ClientConn) Name() string {
	cc.mu.Lock()
	name := cc.name
	cc.mu.Unlock()
	return name
}

// SetHeartBeat sets the heart beats of client connection.
func (cc *ClientConn) SetHeartBeat(heart int64) {
	cc.mu.Lock()
	cc.heart = heart
	cc.mu.Unlock()
}

// HeartBeat gets the heart beats of client connection.
func (cc *ClientConn) HeartBeat() int64 {
	cc.mu.Lock()
	heart := cc.heart
	cc.mu.Unlock()
	return heart
}

// SetContextValue sets extra data to client connection.
func (cc *ClientConn) SetContextValue(k, v interface{}) {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	cc.ctx = context.WithValue(cc.ctx, k, v)
}

// ContextValue gets extra data from client connection.
func (cc *ClientConn) ContextValue(k interface{}) interface{} {
	cc.mu.Lock()
	defer cc.mu.Unlock()
	return cc.ctx.Value(k)
}

// Start starts the client connection, creating go-routines for reading,
// writing and handlng.
func (cc *ClientConn) Start() {
	bylog.Info("conn start, <%v -> %v>\n", cc.rawConn.LocalAddr(), cc.rawConn.RemoteAddr())
	onConnect := cc.opts.onConnect
	if onConnect != nil {
		onConnect(cc)
	}

	loopers := []func(WriteCloser, *sync.WaitGroup){readLoop, writeLoop, handleLoop}
	for _, l := range loopers {
		looper := l
		cc.wg.Add(1)
		go looper(cc, cc.wg)
	}
}

// Close gracefully closes the client connection. It blocked until all sub
// go-routines are completed and returned.
func (cc *ClientConn) Close() {
	cc.once.Do(func() {
		bylog.Info("conn close gracefully, <%v -> %v>\n", cc.rawConn.LocalAddr(), cc.rawConn.RemoteAddr())

		// callback on close
		onClose := cc.opts.onClose
		if onClose != nil {
			onClose(cc)
		}

		// close net.Conn, any blocked read or write operation will be unblocked and
		// return errors.
		cc.rawConn.Close()

		// cancel readLoop, writeLoop and handleLoop go-routines.
		cc.mu.Lock()
		cc.cancel()
		cc.pending = nil
		cc.mu.Unlock()

		// stop timer
		//cc.timing.Stop()

		// wait until all go-routines exited.
		cc.wg.Wait()

		// close all channels.
		close(cc.sendCh)
		close(cc.handlerCh)

		// cc.once is a *sync.Once. After reconnect() returned, cc.once will point
		// to a newly-allocated one while other go-routines such as readLoop,
		// writeLoop and handleLoop blocking on the old *sync.Once continue to
		// execute Close() (and of course do nothing because of sync.Once).
		// NOTE that it will cause an "unlock of unlocked mutex" error if cc.once is
		// a sync.Once struct, because "defer o.m.Unlock()" in sync.Once.Do() will
		// be performed on an unlocked mutex(the newly-allocated one noticed above)
		if cc.opts.reconnect {
			cc.reconnect()
		}
	})
}

// reconnect reconnects and returns a new *ClientConn.
func (cc *ClientConn) reconnect() {
	var c net.Conn
	var err error



	lAddr,err:=net.ResolveUDPAddr("udp",cc.laddr)
	rAddr,err:=net.ResolveUDPAddr("udp",cc.addr)


	//net.ResolveUDPAddr("udp","127.0.0.1:12345")
	c, err = net.DialUDP("udp", lAddr,rAddr)
	//
	//c, err = net.DialUDP("udp", cc.addr)
	if err != nil {
		bylog.Fatal("net dial error", err)
	}

	// copy the newly-created *ClientConn to cc, so after
	// reconnect returned cc will be updated to new one.
	*cc = *newClientConnWithOptions(cc.netid, c, cc.opts)
	cc.Start()
}

// Write writes a message to the client.
func (cc *ClientConn) Write(message Message) error {
	return asyncWrite(cc, message)
}

// RemoteAddr returns the remote address of server connection.
func (cc *ClientConn) RemoteAddr() net.Addr {
	return cc.rawConn.RemoteAddr()
}

// LocalAddr returns the local address of server connection.
func (cc *ClientConn) LocalAddr() net.Addr {
	return cc.rawConn.LocalAddr()
}

func asyncWrite(c interface{}, m Message) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = ErrServerClosed
		}
	}()

	var (
		pkt    []byte
		sendCh chan []byte
	)
	switch c := c.(type) {

	case *ClientConn:
		pkt, err = c.opts.codec.Encode(m)
		sendCh = c.sendCh
	}


	if err != nil {
		bylog.Error("asyncWrite error %v\n", err)
		return
	}

	select {
	case sendCh <- pkt:
		err = nil
	default:
		err = ErrWouldBlock
	}
	return
}


/* readLoop() blocking read from connection, deserialize bytes into message,
then find corresponding handler, put it into channel */
func readLoop(c WriteCloser, wg *sync.WaitGroup) {
	var (
		rawConn          net.Conn
		codec            Codec
		cDone            <-chan struct{}
		//sDone            <-chan struct{}
		setHeartBeatFunc func(int64)
		onMessage        onMessageFunc
		handlerCh        chan MessageHandler
		msg              Message
		err              error
	)

	switch c := c.(type) {

	case *ClientConn:
		rawConn = c.rawConn
		codec = c.opts.codec
		cDone = c.ctx.Done()
		//sDone = nil
		setHeartBeatFunc = c.SetHeartBeat
		onMessage = c.opts.onMessage
		handlerCh = c.handlerCh
	}

	defer func() {
		if p := recover(); p != nil {
			bylog.Error("panics: %v\n", p)
		}
		wg.Done()
		bylog.Debug("readLoop go-routine exited")
		c.Close()
	}()

	for {
		select {
		case <-cDone: // connection closed
			bylog.Debug("receiving cancel signal from conn")
			return
		default:
			msg, err = codec.Decode(rawConn)
			if err != nil {
				bylog.Error("error decoding message %v\n", err)
				continue

				//UDP解码错误的话，直接跳过就可以了。
				//if _, ok := err.(ErrUndefined); ok {
				//	// update heart beats
				//	setHeartBeatFunc(time.Now().UnixNano())
				//	continue
				//}
				//return
			}
			setHeartBeatFunc(time.Now().UnixNano())
			handler := GetHandlerFunc(msg.MessageNumber())
			if handler == nil {
				if onMessage != nil {
					bylog.Info("message %d call onMessage()\n", msg.MessageNumber())
					onMessage(msg, c.(WriteCloser))
				} else {
					bylog.Warn("no handler or onMessage() found for message %d\n", msg.MessageNumber())
				}
				continue
			}
			handlerCh <- MessageHandler{msg, handler}
		}
	}
}

/* writeLoop() receive message from channel, serialize it into bytes,
then blocking write into connection */
func writeLoop(c WriteCloser, wg *sync.WaitGroup) {
	var (
		rawConn net.Conn
		sendCh  chan []byte
		cDone   <-chan struct{}
		//sDone   <-chan struct{}
		pkt     []byte
		err     error
	)

	switch c := c.(type) {

	case *ClientConn:
		rawConn = c.rawConn
		sendCh = c.sendCh
		cDone = c.ctx.Done()
		//sDone = nil
	}

	defer func() {
		if p := recover(); p != nil {
			bylog.Error("panics: %v\n", p)
		}
		// drain all pending messages before exit
	OuterFor:
		for {
			select {
			case pkt = <-sendCh:
				if pkt != nil {
					if _, err = rawConn.Write(pkt); err != nil {
						bylog.Error("error writing data %v\n", err)
					}
				}
			default:
				break OuterFor
			}
		}
		wg.Done()
		bylog.Debug("writeLoop go-routine exited")
		c.Close()
	}()

	for {
		select {
		case <-cDone: // connection closed
			bylog.Debug("receiving cancel signal from conn")
			return
		//case <-sDone: // server closed
		//	holmes.Debugln("receiving cancel signal from server")
		//	return
		case pkt = <-sendCh:
			if pkt != nil {
				if _, err = rawConn.Write(pkt); err != nil {
					bylog.Error("error writing data %v\n", err)
					return
				}
			}
		}
	}
}

// handleLoop() - put handler or timeout callback into worker go-routines
func handleLoop(c WriteCloser, wg *sync.WaitGroup) {
	var (
		cDone        <-chan struct{}
		//sDone        <-chan struct{}
		timerCh      chan *OnTimeOut
		handlerCh    chan MessageHandler
		netID        int64
		ctx          context.Context
		askForWorker bool
		err          error
	)

	switch c := c.(type) {

	case *ClientConn:
		cDone = c.ctx.Done()
		timerCh = c.timing.timeOutChan
		handlerCh = c.handlerCh
		netID = c.netid
		ctx = c.ctx
	}

	defer func() {
		if p := recover(); p != nil {
			bylog.Error("panics: %v\n", p)
		}
		wg.Done()
		bylog.Debug("handleLoop go-routine exited")
		c.Close()
	}()

	for {
		select {
		case <-cDone: // connectin closed
			bylog.Warn("receiving cancel signal from conn")
			return
		case msgHandler := <-handlerCh:
			msg, handler := msgHandler.message, msgHandler.handler
			if handler != nil {
				if askForWorker {
					err = WorkerPoolInstance().Put(netID, func() {
						handler(NewContextWithNetID(NewContextWithMessage(ctx, msg), netID), c)
					})
					if err != nil {
						bylog.Error("%s",err)
					}
					addTotalHandle()
				} else {
					handler(NewContextWithNetID(NewContextWithMessage(ctx, msg), netID), c)
				}
			}
		case timeout := <-timerCh:
			if timeout != nil {
				timeoutNetID := NetIDFromContext(timeout.Ctx)
				if timeoutNetID != netID {
					bylog.Error("timeout net %d, conn net %d, mismatched!\n", timeoutNetID, netID)
				}
				if askForWorker {
					err = WorkerPoolInstance().Put(netID, func() {
						timeout.Callback(time.Now(), c.(WriteCloser))
					})
					if err != nil {
						bylog.Error("%s",err)
					}
				} else {
					timeout.Callback(time.Now(), c.(WriteCloser))
				}
			}
		}
	}
}
