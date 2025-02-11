// Copyright 2020 The goftp Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package integrations

import (
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"SimPro/pkg/goftp.io/server/v2"
	"SimPro/pkg/goftp.io/server/v2/driver/minio"

	"github.com/jlaffaye/ftp"
	"github.com/stretchr/testify/assert"
)

func TestDriver(t *testing.T) {
	endpoint := os.Getenv("MINIO_SERVER_ENDPOINT")
	if endpoint == "" {
		t.Skip()
		return
	}
	accessKeyID := os.Getenv("MINIO_SERVER_ACCESS_KEY_ID")
	secretKey := os.Getenv("MINIO_SERVER_SECRET_KEY")
	location := os.Getenv("MINIO_SERVER_LOCATION")
	bucket := os.Getenv("MINIO_SERVER_BUCKET")
	useSSL, _ := strconv.ParseBool(os.Getenv("MINIO_SERVER_USE_SSL"))

	minioDriver, err := minio.NewDriver(endpoint, accessKeyID, secretKey, location, bucket, useSSL)
	assert.NoError(t, err)
	opt := &server.Options{
		Name:   "test ftpd",
		Driver: minioDriver,
		Port:   2120,
		Auth: &server.SimpleAuth{
			Name:     "admin",
			Password: "admin",
		},
		Perm:   server.NewSimplePerm("root", "root"),
		Logger: new(server.DiscardLogger),
	}

	runServer(t, opt, nil, func() {
		// Give server 0.5 seconds to get to the listening state
		timeout := time.NewTimer(time.Millisecond * 500)
		for {
			f, err := ftp.Connect("localhost:2120")
			if err != nil && len(timeout.C) == 0 { // Retry errors
				continue
			}

			assert.NoError(t, err)
			assert.NotNil(t, f)

			assert.NoError(t, f.Login("admin", "admin"))
			assert.Error(t, f.Login("admin", ""))

			curDir, err := f.CurrentDir()
			assert.NoError(t, err)
			assert.EqualValues(t, "/", curDir)

			err = f.RemoveDir("/")
			assert.Error(t, err)

			var content = `test`
			assert.NoError(t, f.Stor("server_test.go", strings.NewReader(content)))

			r, err := f.Retr("server_test.go")
			assert.NoError(t, err)

			buf, err := ioutil.ReadAll(r)
			assert.NoError(t, err)
			r.Close()

			assert.EqualValues(t, content, buf)

			entries, err := f.List("/")
			assert.NoError(t, err)
			assert.EqualValues(t, 1, len(entries))
			assert.EqualValues(t, "server_test.go", entries[0].Name)
			assert.EqualValues(t, ftp.EntryTypeFile, entries[0].Type)
			assert.EqualValues(t, len(buf), entries[0].Size)

			size, err := f.FileSize("/server_test.go")
			assert.NoError(t, err)
			assert.EqualValues(t, 4, size)

			assert.NoError(t, f.Delete("/server_test.go"))

			entries, err = f.List("/")
			assert.NoError(t, err)
			assert.EqualValues(t, 0, len(entries))

			assert.NoError(t, f.Stor("server_test2.go", strings.NewReader(content)))

			err = f.RemoveDir("/")
			assert.Error(t, err)

			entries, err = f.List("/")
			assert.NoError(t, err)
			assert.EqualValues(t, 1, len(entries))

			assert.NoError(t, f.Delete("/server_test2.go"))

			assert.NoError(t, f.Stor("server_test3.go", strings.NewReader(content)))

			err = f.Rename("/server_test3.go", "/test.go")
			assert.NoError(t, err)

			entries, err = f.List("/")
			assert.NoError(t, err)
			assert.EqualValues(t, 1, len(entries))
			assert.EqualValues(t, "test.go", entries[0].Name)
			assert.EqualValues(t, 4, entries[0].Size)
			assert.EqualValues(t, ftp.EntryTypeFile, entries[0].Type)

			err = f.MakeDir("/src")
			assert.NoError(t, err)

			err = f.ChangeDir("/src")
			assert.NoError(t, err)

			curDir, err = f.CurrentDir()
			assert.NoError(t, err)
			assert.EqualValues(t, "/src", curDir)

			err = f.MakeDir("/new/1/2")
			assert.NoError(t, err)

			entries, err = f.List("/new/1")
			assert.NoError(t, err)
			assert.EqualValues(t, 1, len(entries))
			assert.EqualValues(t, "2/", entries[0].Name)
			assert.EqualValues(t, 0, entries[0].Size)
			assert.EqualValues(t, ftp.EntryTypeFolder, entries[0].Type)

			assert.NoError(t, f.Stor("/test/1/2/server_test3.go", strings.NewReader(content)))

			r, err = f.RetrFrom("/test/1/2/server_test3.go", 2)
			assert.NoError(t, err)

			buf, err = ioutil.ReadAll(r)
			r.Close()
			assert.NoError(t, err)
			assert.EqualValues(t, "st", string(buf))

			curDir, err = f.CurrentDir()
			assert.NoError(t, err)
			assert.EqualValues(t, "/src", curDir)

			assert.NoError(t, f.Stor("server_test.go", strings.NewReader(content)))

			r, err = f.Retr("/src/server_test.go")
			assert.NoError(t, err)

			buf, err = ioutil.ReadAll(r)
			r.Close()
			assert.NoError(t, err)
			assert.EqualValues(t, "test", string(buf))

			break
		}
	})
}
