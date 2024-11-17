package main

import (
	photoutil "github.com/tanner-caffrey/gwynnie-back-paws/photoutil"
)

func main() {
	photoutil.UpdatePhotosInteractive(photoutil.DefaultInteractiveConfig())
}
