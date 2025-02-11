// +ignore

package main

import (
	"log"

	"SimPro/pkg/goftp.io/server/v2"
	"SimPro/pkg/goftp.io/server/v2/driver/file"
)

func main() {
	driver, err := file.NewDriver("./")
	if err != nil {
		log.Fatal(err)
	}

	s, err := server.NewServer(&server.Options{
		Driver: driver,
		Auth: &server.SimpleAuth{
			Name:     "admin",
			Password: "admin",
		},
		Perm:      server.NewSimplePerm("root", "root"),
		RateLimit: 1000000, // 1MB/s limit
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
