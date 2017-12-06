package main

import (
	"fmt"
	"os"
)

func main() {
	cacheFolder := `C:\game\World of Warcraft\Cache`

	fmt.Print("Deleting cache folder...")
	err := os.RemoveAll(cacheFolder)
	if err != nil {
		fmt.Println("failed")
	}

	fmt.Println("ok")
	fmt.Scanln()
}
