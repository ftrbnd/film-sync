package main

import "github.com/ftrbnd/film-sync/internal/files"

func main() {
	filePath := files.OpenZip("wetransfer.zip")
	files.ConvertToPNG(filePath, "tif")
}
