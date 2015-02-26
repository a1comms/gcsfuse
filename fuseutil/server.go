// Copyright 2015 Google Inc. All Rights Reserved.
// Author: jacobsa@google.com (Aaron Jacobs)

package fuseutil

import (
	"fmt"
	"io"
	"log"

	"bazil.org/fuse"
)

// An object that terminates one end of the userspace <-> FUSE VFS connection.
type server struct {
	logger *log.Logger
	fs     FileSystem
}

// Create a server that relays requests to the supplied file system.
func newServer(fs FileSystem) (s *server, err error) {
	s = &server{
		logger: getLogger(),
		fs:     fs,
	}

	return
}

// Serve the fuse connection by repeatedly reading requests from the supplied
// FUSE connection, responding as dictated by the file system. Return when the
// connection is closed or an unexpected error occurs.
func (s *server) Serve(c *fuse.Conn) (err error) {
	// Read a message at a time, dispatching to goroutines doing the actual
	// processing.
	for {
		var fuseReq fuse.Request
		fuseReq, err = c.ReadRequest()

		// ReadRequest returns EOF when the connection has been closed.
		//
		// TODO(jacobsa): Remove this and verify it's actually needed.
		if err == io.EOF {
			err = nil
			return
		}

		// Otherwise, forward on errors.
		if err != nil {
			err = fmt.Errorf("Conn.ReadRequest: %v", err)
			return
		}

		go s.handleFuseRequest(fuseReq)
	}
}

func (s *server) handleFuseRequest(fuseReq fuse.Request) {
	// Log the request.
	s.logger.Println("Received:", fuseReq)

	// Attempt to handle it.
	switch typed := fuseReq.(type) {
	case *fuse.InitRequest:
		// Responding to this is required to make mounting work, at least on OS X.
		// We don't currently expose the capability for the file system to
		// intercept this.
		fuseResp := &fuse.InitResponse{}
		s.logger.Println("Responding:", fuseResp)
		typed.Respond(fuseResp)

	default:
		s.logger.Println("Unhandled type. Returning ENOSYS.")
		typed.RespondError(ENOSYS)
	}
}
