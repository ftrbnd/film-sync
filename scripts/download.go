package main

import "github.com/ftrbnd/film-sync/internal/files"

func main() {
	dst := "output"

	files.OpenZip("wetransfer.zip", dst)
	files.ConvertToPNG("tif", dst)
}
