package main

import (
	"testing"

	"github.com/verdverm/frisby"
)

func init() {
	frisby.Global.PrintProgressDot = false
}

func Test_ListPosts(*testing.T) {
	frisby.Create("Get Posts list").
		Get("http://localhost:8000/posts").
		Send().
		ExpectStatus(200).
		PrintReport()
}
