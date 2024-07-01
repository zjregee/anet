package anet

import (
	"errors"
	"context"
	"sync/atomic"
)

type callbackNode struct {
	fn   CloseCallback
	pre *callbackNode
}

type onEvent struct {
	onConnectCallback atomic.Value
	onRequestCallback atomic.Value
	closeCallbacks    atomic.Value
}

func (c *connection) SetOnConnect(onConnect OnConnect) {
	if onConnect != nil {
		c.onConnectCallback.Store(onConnect)
	}
}

func (c *connection) SetOnRequest(onRequest OnRequest) {
	if onRequest != nil {
		c.onRequestCallback.Store(onRequest)
	}
}

func (c *connection) AddCloseCallback(callback CloseCallback) {
	if callback == nil {
		return
	}
	node := &callbackNode{}
	node.fn = callback
	if pre := c.closeCallbacks.Load(); pre != nil {
		node.pre = pre.(*callbackNode)
	}
	c.closeCallbacks.Store(node)
}

func (c *connection) onPrepare() {
	if c.opts != nil {
		c.SetOnConnect(c.opts.onConnect)
		c.SetOnRequest(c.opts.onRequest)
	}
	c.context = context.Background()
	c.onConnect()
	log.Infof("[conn %s] accepted", c.id)
	go c.readLoop()
	go c.writeLoop()
}

func (c *connection) onConnect() {
	onConnect, _ := c.onConnectCallback.Load().(OnConnect)
	if onConnect == nil {
		atomic.StoreInt32(&c.state, 1)
		return
	}
	onRequest, _ := c.onRequestCallback.Load().(OnRequest)
	c.onProcess(
		func(c *connection) bool {
			return atomic.LoadInt32(&c.state) == 0
		},
		func(c *connection) {
			c.context = onConnect(c.context, c)
			// execute onRequest once here avoid onConnect take too long time
			if onRequest != nil && c.inputBuffer.Len() > 0 {
				err := onRequest(c.context, c)
				if err != nil {
					log.Fatal("error occurred while executing OnRequest")
				}
			}
			atomic.StoreInt32(&c.state, 1)
		},
	)
}

func (c *connection) onRequest() bool {
	onRequest, _ := c.onRequestCallback.Load().(OnRequest)
	if onRequest == nil {
		return false
	}
	onConnect, _ := c.onConnectCallback.Load().(OnConnect)
	if onConnect != nil && atomic.LoadInt32(&c.state) == 0 {
		// onConnect should finished first and wait for onConnect to retrigger
		return false
	}
	processed := c.onProcess(
		func(c *connection) bool {
			return c.inputBuffer.Len() > 0
		},
		func(c *connection) {
			defer c.unlock(processing)
			log.Infof("[conn %s] executes onRequest once", c.id)
			err := onRequest(c.context, c)
			if err != nil {
				log.Fatal("error occurred while executing OnRequest")
			}
		},
	)
	return processed
}

func (c *connection) onProcess(isProcessable func(c *connection) bool, process func(c *connection)) bool {
	if process == nil {
		return false
	}
	if !c.lock(processing) {
		return false
	}
	go func() {
		for {
			if isProcessable(c) {
				process(c)
			}
		}
	}()
	// go func() {
	// 	panicked := true
	// 	defer func() {
	// 		if panicked {
	// 			c.unlock(processing)
	// 			c.onClose()
	// 		}
	// 	}()
	// 	for {
	// 		if isProcessable(c) {
	// 			process(c)
	// 		}
	// 		var closedBy int32
	// 		for {
	// 			closedBy = c.status(closing)
	// 			if closedBy != none || !isProcessable(c) {
	// 				break
	// 			}
	// 			process(c)
	// 		}
	// 		if closedBy == ring {
	// 			_ = c.closeCallback()
	// 			panicked = false
	// 			return
	// 		}
	// 		c.unlock(processing)

	// 		if isProcessable(c) && c.lock(processing) {
	// 			continue
	// 		}
	// 		panicked = false
	// 		return
	// 	}
	// }()
	return true
}

func (c *connection) closeCallback() error {
	latest := c.closeCallbacks.Load()
	if latest == nil {
		return nil
	}
	var err error
	var errs []error
	for callback := latest.(*callbackNode); callback != nil; callback = callback.pre {
		err = callback.fn(c)
		errs = append(errs, err)
	}
	err = errors.Join(errs...)
	if err != nil {
		log.Fatal("error occurred while executing CloseCallBacks")
	}
	return err
}
