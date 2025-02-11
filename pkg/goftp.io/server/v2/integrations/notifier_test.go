// Copyright 2020 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"SimPro/pkg/goftp.io/server/v2"
	"SimPro/pkg/goftp.io/server/v2/driver/file"

	"github.com/jlaffaye/ftp"
	"github.com/stretchr/testify/assert"
)

var (
	_ server.Notifier = &mockNotifier{}
)

type mockNotifier struct {
	actions []string
	lock    sync.Mutex
}

func (m *mockNotifier) BeforeLoginUser(ctx *server.Context, userName string) {
	m.lock.Lock()
	m.actions = append(m.actions, "BeforeLoginUser")
	m.lock.Unlock()
}
func (m *mockNotifier) BeforePutFile(ctx *server.Context, dstPath string) {
	m.lock.Lock()
	m.actions = append(m.actions, "BeforePutFile")
	m.lock.Unlock()
}
func (m *mockNotifier) BeforeDeleteFile(ctx *server.Context, dstPath string) {
	m.lock.Lock()
	m.actions = append(m.actions, "BeforeDeleteFile")
	m.lock.Unlock()
}
func (m *mockNotifier) BeforeChangeCurDir(ctx *server.Context, oldCurDir, newCurDir string) {
	m.lock.Lock()
	m.actions = append(m.actions, "BeforeChangeCurDir")
	m.lock.Unlock()
}
func (m *mockNotifier) BeforeCreateDir(ctx *server.Context, dstPath string) {
	m.lock.Lock()
	m.actions = append(m.actions, "BeforeCreateDir")
	m.lock.Unlock()
}
func (m *mockNotifier) BeforeDeleteDir(ctx *server.Context, dstPath string) {
	m.lock.Lock()
	m.actions = append(m.actions, "BeforeDeleteDir")
	m.lock.Unlock()
}
func (m *mockNotifier) BeforeDownloadFile(ctx *server.Context, dstPath string) {
	m.lock.Lock()
	m.actions = append(m.actions, "BeforeDownloadFile")
	m.lock.Unlock()
}
func (m *mockNotifier) AfterUserLogin(ctx *server.Context, userName, password string, passMatched bool, err error) {
	m.lock.Lock()
	m.actions = append(m.actions, "AfterUserLogin")
	m.lock.Unlock()
}
func (m *mockNotifier) AfterFilePut(ctx *server.Context, dstPath string, size int64, err error) {
	m.lock.Lock()
	m.actions = append(m.actions, "AfterFilePut")
	m.lock.Unlock()
}
func (m *mockNotifier) AfterFileDeleted(ctx *server.Context, dstPath string, err error) {
	m.lock.Lock()
	m.actions = append(m.actions, "AfterFileDeleted")
	m.lock.Unlock()
}
func (m *mockNotifier) AfterCurDirChanged(ctx *server.Context, oldCurDir, newCurDir string, err error) {
	m.lock.Lock()
	m.actions = append(m.actions, "AfterCurDirChanged")
	m.lock.Unlock()
}
func (m *mockNotifier) AfterDirCreated(ctx *server.Context, dstPath string, err error) {
	m.lock.Lock()
	m.actions = append(m.actions, "AfterDirCreated")
	m.lock.Unlock()
}
func (m *mockNotifier) AfterDirDeleted(ctx *server.Context, dstPath string, err error) {
	m.lock.Lock()
	m.actions = append(m.actions, "AfterDirDeleted")
	m.lock.Unlock()
}
func (m *mockNotifier) AfterFileDownloaded(ctx *server.Context, dstPath string, size int64, err error) {
	m.lock.Lock()
	m.actions = append(m.actions, "AfterFileDownloaded")
	m.lock.Unlock()
}

func assetMockNotifier(t *testing.T, mock *mockNotifier, lastActions []string) {
	if len(lastActions) == 0 {
		return
	}
	mock.lock.Lock()
	assert.EqualValues(t, lastActions, mock.actions[len(mock.actions)-len(lastActions):])
	mock.lock.Unlock()
}

func TestNotification(t *testing.T) {
	err := os.MkdirAll("./testdata", os.ModePerm)
	assert.NoError(t, err)

	var perm = server.NewSimplePerm("test", "test")
	driver, err := file.NewDriver("./testdata")
	assert.NoError(t, err)

	opt := &server.Options{
		Name:   "test ftpd",
		Driver: driver,
		Port:   2121,
		Auth: &server.SimpleAuth{
			Name:     "admin",
			Password: "admin",
		},
		Perm:   perm,
		Logger: new(server.DiscardLogger),
	}

	mock := &mockNotifier{}

	runServer(t, opt, []server.Notifier{mock}, func() {
		// Give server 0.5 seconds to get to the listening state
		timeout := time.NewTimer(time.Millisecond * 500)

		for {
			f, err := ftp.Connect("localhost:2121")
			if err != nil && len(timeout.C) == 0 { // Retry errors
				continue
			}
			assert.NoError(t, err)

			assert.NoError(t, f.Login("admin", "admin"))
			assetMockNotifier(t, mock, []string{"BeforeLoginUser", "AfterUserLogin"})

			assert.Error(t, f.Login("admin", "1111"))
			assetMockNotifier(t, mock, []string{"BeforeLoginUser", "AfterUserLogin"})

			var content = `test`
			assert.NoError(t, f.Stor("server_test.go", strings.NewReader(content)))
			assetMockNotifier(t, mock, []string{"BeforePutFile", "AfterFilePut"})

			r, err := f.RetrFrom("/server_test.go", 2)
			assert.NoError(t, err)

			buf, err := ioutil.ReadAll(r)
			r.Close()
			assert.NoError(t, err)
			assert.EqualValues(t, "st", string(buf))
			assetMockNotifier(t, mock, []string{"BeforeDownloadFile", "AfterFileDownloaded"})

			err = f.Rename("/server_test.go", "/test.go")
			assert.NoError(t, err)

			err = f.MakeDir("/src")
			assert.NoError(t, err)
			assetMockNotifier(t, mock, []string{"BeforeCreateDir", "AfterDirCreated"})

			err = f.Delete("/test.go")
			assert.NoError(t, err)
			assetMockNotifier(t, mock, []string{"BeforeDeleteFile", "AfterFileDeleted"})

			err = f.ChangeDir("/src")
			assert.NoError(t, err)
			assetMockNotifier(t, mock, []string{"BeforeChangeCurDir", "AfterCurDirChanged"})

			err = f.RemoveDir("/src")
			assert.NoError(t, err)
			assetMockNotifier(t, mock, []string{"BeforeDeleteDir", "AfterDirDeleted"})

			err = f.Quit()
			assert.NoError(t, err)

			break
		}
	})
}
