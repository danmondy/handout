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

	fmt.Println(path)
	pathdir := filepath.Dir(path)
	for _, dir := range u.Directories {
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

//TODO: this function is just a placeholder
func ValidUser(u User) bool {
	if u.Name == "invalid" {
		return false
	}
	return true
}

//-----------------------
