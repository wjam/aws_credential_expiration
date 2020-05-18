// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
)

var file = `//+build linux darwin

package main

var %s = []byte{%s
}
`

func main() {
	if err := generate("green.png", "greeniconunix.go", "greenIcon"); err != nil {
		panic(err)
	}

	if err := generate("red.png", "rediconunix.go", "redIcon"); err != nil {
		panic(err)
	}

	if err := generate("amber.png", "ambericonunix.go", "amberIcon"); err != nil {
		panic(err)
	}
}

func generate(source string, output string, name string) error {
	content, err := ioutil.ReadFile(source)
	if err != nil {
		return err
	}

	arrayContent := ""
	for i, b := range content {
		if i%12 == 0 {
			arrayContent += "\n\t"
		} else {
			arrayContent += " "
		}
		arrayContent += fmt.Sprintf("0x%02x,", b)
	}

	err = ioutil.WriteFile(output, []byte(fmt.Sprintf(file, name, arrayContent)), 0644)
	if err != nil {
		return err
	}

	return nil
}
