package main

import (
	"desc/base"
	"desc/services/youtube"
)

func main() {

	b := base.LoadBase()
	defer b.Kill()

	youtube.RunTasks(b)
}
