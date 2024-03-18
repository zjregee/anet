package anet

import (
	"context"
	"errors"
	"sync/atomic"
)

type callbackNode struct {
	fn  CloseCallback
	pre *callbackNode
}

type onEvent struct {
	ctx               context.Context
	onConnectCallback atomic.Value
	onRequestCallback atomic.Value
	closeCallbacks    atomic.Value
}

func (c *connection) SetOnConnect(onConnect OnConnect) error {
	if onConnect != nil {
		c.onConnectCallback.Store(onConnect)
	}
	return nil
}

func (c *connection) SetOnRequest(onRequest OnRequest) error {
	if onRequest != nil {
		c.onRequestCallback.Store(onRequest)
	}
	return nil
}

func (c *connection) AddCloseCallback(callback CloseCallback) error {
	if callback == nil {
		return nil
	}
	cb := &callbackNode{}
	cb.fn = callback
	if pre := c.closeCallbacks.Load(); pre != nil {
		cb.pre = pre.(*callbackNode)
	}
	c.closeCallbacks.Store(cb)
	return nil
}

func (c *connection) onConnect() {
	onConnect, ok := c.onConnectCallback.Load().(OnConnect)
	if !ok {
		return
	}
	var connected int32
	c.onProcess(
		func(c *connection) bool {
			return atomic.LoadInt32(&connected) == 0
		},
		func(c *connection) {
			if atomic.CompareAndSwapInt32(&connected, 0, 1) {
				c.ctx = onConnect(c.ctx, c)
			}
		},
	)
}

func (c *connection) onRequest() bool {
	onRequest, ok := c.onRequestCallback.Load().(OnRequest)
	if !ok {
		return true
	}
	processed := c.onProcess(
		func(c *connection) bool {
			return len(c.inputBuffer) > 0
		},
		func(c *connection) {
			_ = onRequest(c.ctx, c)
		},
	)
	return !processed
}

func (c *connection) onHup() error {
	return nil
}

func (c *connection) onClose() error {
	if c.closeBy(user) {
		c.triggerRead(errors.New("closed by user"))
		c.triggerWrite(errors.New("closed by user"))
		c.closeCallback(true, true)
		return nil
	}
	return c.closeCallback(true, false)
}

func (c *connection) onProcess(isProcessable func(c *connection) bool, process func(c *connection)) bool {
	if process == nil {
		return false
	}
	if !c.lock(processing) {
		return false
	}
	go func() {
		panicked := true
		defer func() {
			c.unlock(processing)
			if panicked {
				c.closeCallback(false, false)
			}
		}()
		for {
			if isProcessable(c) {
				process(c)
			}
			closedBy := none
			for {
				closedBy = c.status(closing)
				if closedBy != none || !isProcessable(c) {
					break
				}
				process(c)
			}
			if closedBy != none {
				needDetach := closedBy == user
				c.closeCallback(false, needDetach)
				panicked = false
				return
			}
			if isProcessable(c) {
				continue
			}
			panicked = false;
			return
		}
	}()
	return true
}

func (c *connection) closeCallback(needLock bool, needDetach bool) error {
	// ensure closeCallback and onRequest will not be executed concurrently
	if needLock && !c.lock(processing) {
		return nil
	}
	// if connection is registered, should be detached from ring
	if needDetach {
		return nil
	}
	latest := c.closeCallbacks.Load()
	if latest == nil {
		return nil
	}
	for callback := latest.(*callbackNode); callback != nil; callback = callback.pre {
		callback.fn(c)
	}
	return nil
}
