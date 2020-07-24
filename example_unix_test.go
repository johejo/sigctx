// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build aix darwin dragonfly freebsd js,wasm linux netbsd openbsd solaris

package sigctx_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/johejo/sigctx"
)

// This example passes a context with a signal to tell a blocking function that
// it should abandon its work after a signal is received.
func ExampleNotifyContext() {
	// Pass a context with a timeout to tell a blocking function that it
	// should abandon its work after the timeout elapses.
	ctx, stop := sigctx.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		log.Fatal(err)
	}

	// On a Unix-like system, pressing Ctrl+C on a keyboard sends the SIGINT signal to the process of the program in execution.
	// This example simulates it by sending a SIGINT signal to itself.
	p.Signal(os.Interrupt)
	if err != nil {
		log.Fatal(err)
	}

	select {
	case <-time.After(time.Second):
		fmt.Println("missed signal")
	case <-ctx.Done():
		fmt.Println(ctx.Err()) // prints "context canceled"
		stop()                 // stop receiving signal notifications as soon as possible.
	}

	// Output:
	// context canceled
}
