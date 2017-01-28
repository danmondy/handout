package main

import (
	"fmt"
	"path/filepath"
)

//============USER============
type User struct {
	Name        string
	Pword       string
	Directories []string
	Files       []string
}

func (u User) CanEditFile(path string) bool {
	/*f, err := os.Stat(path)
	if err != nil {
		//file does not exits
		return false
	}*/

	fmt.Println("PATHOF: " + path)
	pathdir := filepath.Dir(path) + "/"
	fmt.Println("DIROF: " + pathdir )
	for _, dir := range u.Directories {
		fmt.Println("Trying: " + dir)
		if dir == pathdir {
			return true
		}
	}
	for _, uf := range u.Files {
		if uf == path {
			return true
		}
	}
	return false
}

//-----------------------
