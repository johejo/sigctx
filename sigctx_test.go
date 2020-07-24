// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd linux netbsd openbsd solarispackage sigctx

package sigctx

import (
	"context"
	"fmt"
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
		t.Errorf("c.String() = %q want prefix %q", got, prefix)
	}

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	select {
	case <-c.Done():
		if got := c.Err(); got != context.Canceled {
			t.Errorf("c.Err() = %q want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("timed out waiting for context to be done after SIGINT")
	}
}

func TestNotifyContextCanceled(t *testing.T) {
	c, stop := NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	if got, prefix := fmt.Sprint(c), "signal.NotifyContext("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q want prefix %q", got, prefix)
	}

	stop()
	select {
	case <-c.Done():
		if got := c.Err(); got != context.Canceled {
			t.Errorf("c.Err() = %q want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("timed out waiting for context to be done after SIGINT")
	}
}

func TestNotifyContextSimultaneousNotifications(t *testing.T) {
	c, stop := NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	if got, prefix := fmt.Sprint(c), "signal.NotifyContext("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q want prefix %q", got, prefix)
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
			t.Errorf("c.Err() = %q want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("expected context to be canceled")
	}
}

func TestNotifyContextSimultaneousStop(t *testing.T) {
	c, stop := NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	if got, prefix := fmt.Sprint(c), "signal.NotifyContext("; !strings.HasPrefix(got, prefix) {
		t.Errorf("c.String() = %q want prefix %q", got, prefix)
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
			t.Errorf("c.Err() = %q want %q", got, context.Canceled)
		}
	case <-time.After(time.Second):
		t.Errorf("expected context to be canceled")
	}
}
