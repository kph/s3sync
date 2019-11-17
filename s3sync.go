// Copyright 2019 Kevin Paul Herbert

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func transferDir(dir string) (err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("ReadDir of %s failed: %w", dir)
	}
	for _, file := range files {
		fn := filepath.Join(dir, file.Name())
		fmt.Printf("Working on %s\n", fn)
		if file.IsDir() {
			err := transferDir(fn)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func main() {
	transferDir(os.Args[1])
}
