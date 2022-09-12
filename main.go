package main

import "simpleGinIm/router"

func main() {
	e := router.Router()
	e.Run(":8080")
}
