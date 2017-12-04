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
		PrintBody()
}

func Test_ListOnePosts(*testing.T) {
	frisby.Create("Get Posts 1").
		Get("http://localhost:8000/posts/1").
		Send().
		ExpectStatus(200).
		PrintBody()
}

func Test_AddPosts(*testing.T) {
	frisby.Create("Add Post").
		Post("http://localhost:8000/posts").
			SetJson( map[string]string {
				"columnone":"test1",
				"columntwo":"test2",
				"columnthree":"test3",
			} ).
		Send().
		ExpectStatus(200).
		PrintBody()
}


func Test_UpdatePosts(*testing.T) {
	frisby.Create("Update Post").
		Put("http://localhost:8000/posts/1").
		SetJson( map[string]string {
		"columnone":"updated1",
		"columntwo":"updated2",
		"columnthree":"updated3",
	} ).
		Send().
		ExpectStatus(200).
		PrintBody()
}

func Test_DeletePosts(*testing.T) {
	frisby.Create("Update Post").
		Delete("http://localhost:8000/posts/4").
		Send().
		ExpectStatus(200).
		PrintBody()
}
