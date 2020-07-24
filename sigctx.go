// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.package sigctx

package sigctx

import (
	"context"
	"os"
	"os/signal"
	"time"
)

// NotifyContext returns a copy of the parent context that is marked done
// (its Done channel is closed) when one of the listed signals arrives,
// when the returned stop function is called, or when the parent context's
// Done channel is closed, whichever happens first.
//
// The stop function unregisters the signal behavior, which, like signal.Reset,
// may restore the default behavior for a given signal. For example, the default
// behavior of a Go program receiving os.Interrupt is to exit. Calling
// NotifyContext(parent, os.Interrupt) will change the behavior to cancel
// the returned context. Future interrupts received will not trigger the default
// (exit) behavior until the returned stop function is called.
//
// The stop function releases resources associated with it, so code should
// call stop as soon as the operations running in this Context complete and
// signals no longer need to be diverted to the context.
func NotifyContext(parent context.Context, signals ...os.Signal) (ctx context.Context, stop context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	c := &signalCtx{
		parent:  ctx,
		cancel:  cancel,
		signals: signals,
	}
	if ctx.Err() == nil {
		c.ch = make(chan os.Signal, 1)
		signal.Notify(c.ch, c.signals...)
		go func() {
			select {
			case <-c.ch:
				c.cancel()
			case <-c.Done():
			}
		}()
	}
	return c, c.stop
}

type signalCtx struct {
	parent context.Context

	cancel  context.CancelFunc
	signals []os.Signal
	ch      chan os.Signal
}

func (c *signalCtx) stop() {
	c.cancel()
	signal.Stop(c.ch)
}

func (c *signalCtx) String() string {
	var buf []byte
	buf = append(buf, "signal.NotifyContext(parent"...)
	if len(c.signals) != 0 {
		buf = append(buf, ", ["...)
		for _, s := range c.signals {
			buf = append(buf, s.String()...)
			buf = append(buf, ' ')
		}
		buf = append(buf, ']')
	}
	buf = append(buf, ')')
	return string(buf)
}

func (c *signalCtx) Deadline() (deadline time.Time, ok bool) {
	return c.parent.Deadline()
}

func (c *signalCtx) Done() <-chan struct{} {
	return c.parent.Done()
}

func (c *signalCtx) Err() error {
	return c.parent.Err()
}

func (c *signalCtx) Value(key interface{}) interface{} {
	return c.parent.Value(key)
}
