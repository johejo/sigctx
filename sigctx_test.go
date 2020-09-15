// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solarispackage sigctx

package sigctx

import (
	"context"
	"fmt"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

func TestNotifyContext(t *testing.T) {
	c, stop := NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	if got, prefix := fmt.Sprint(c), "signal.NotifyContext("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q, want prefix %q", got, prefix)
	}

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	select {
	case <-c.Done():
		if got := c.Err(); got != context.Canceled {
			t.Errorf("c.Err() = %q, want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("timed out waiting for context to be done after SIGINT")
	}
}

func TestNotifyContextStop(t *testing.T) {
	signal.Ignore(syscall.SIGHUP)
	if !signal.Ignored(syscall.SIGHUP) {
		t.Errorf("expected SIGHUP to be ignored when explicitly ignoring it.")
	}

	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	c, stop := NotifyContext(parent, syscall.SIGHUP)
	defer stop()

	// If we're being notified, then the signal should not be ignored.
	if signal.Ignored(syscall.SIGHUP) {
		t.Errorf("expected SIGHUP to not be ignored.")
	}

	if got, prefix := fmt.Sprint(c), "signal.NotifyContext("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q, want prefix %q", got, prefix)
	}

	stop()
	select {
	case <-c.Done():
		if got := c.Err(); got != context.Canceled {
			t.Errorf("c.Err() = %q, want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("timed out waiting for context to be done after calling stop")
	}
}

func TestNotifyContextCancelParent(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	c, stop := NotifyContext(parent, syscall.SIGINT)
	defer stop()

	if got, prefix := fmt.Sprint(c), "signal.NotifyContext("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q, want prefix %q", got, prefix)
	}

	cancelParent()
	select {
	case <-c.Done():
		if got := c.Err(); got != context.Canceled {
			t.Errorf("c.Err() = %q, want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("timed out waiting for parent context to be canceled")
	}
}

func TestNotifyContextPrematureCancelParent(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()

	cancelParent() // Prematurely cancel context before calling NotifyContext.
	c, stop := NotifyContext(parent, syscall.SIGINT)
	defer stop()

	if got, prefix := fmt.Sprint(c), "signal.NotifyContext("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q, want prefix %q", got, prefix)
	}

	select {
	case <-c.Done():
		if got := c.Err(); got != context.Canceled {
			t.Errorf("c.Err() = %q, want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("timed out waiting for parent context to be canceled")
	}
}

func TestNotifyContextSimultaneousNotifications(t *testing.T) {
	c, stop := NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	if got, prefix := fmt.Sprint(c), "signal.NotifyContext("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q, want prefix %q", got, prefix)
	}

	var wg sync.WaitGroup
	n := 10
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			wg.Done()
		}()
	}
	wg.Wait()
	select {
	case <-c.Done():
		if got := c.Err(); got != context.Canceled {
			t.Errorf("c.Err() = %q, want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("expected context to be canceled")
	}
}

func TestNotifyContextSimultaneousStop(t *testing.T) {
	c, stop := NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	if got, prefix := fmt.Sprint(c), "signal.NotifyContext("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q, want prefix %q", got, prefix)
	}

	var wg sync.WaitGroup
	n := 10
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			stop()
			wg.Done()
		}()
	}
	wg.Wait()
	select {
	case <-c.Done():
		if got := c.Err(); got != context.Canceled {
			t.Errorf("c.Err() = %q, want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("expected context to be canceled")
	}
}

func TestNotifyContextStringer(t *testing.T) {
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	c, stop := NotifyContext(parent, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	want := `signal.NotifyContext(context.Background.WithCancel, [hangup interrupt terminated])`
	if got := fmt.Sprint(c); got != want {
		t.Errorf("c.String() = %q, want %q", got, want)
	}
}
