package main

import (
	"desc/base"
)

func main() {

	b := base.LoadBase()
	defer b.Kill()

}
