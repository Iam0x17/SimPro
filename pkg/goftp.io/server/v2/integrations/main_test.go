// Copyright 2018 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"testing"

	"SimPro/pkg/goftp.io/server/v2"

	"github.com/stretchr/testify/assert"
)

func runServer(t *testing.T, opt *server.Options, notifiers []server.Notifier, execute func()) {
	s, err := server.NewServer(opt)
	assert.NoError(t, err)
	for _, notifier := range notifiers {
		s.RegisterNotifer(notifier)
	}
	go func() {
		err := s.ListenAndServe()
		assert.EqualError(t, err, server.ErrServerClosed.Error())
	}()

	execute()

	assert.NoError(t, s.Shutdown())
}
