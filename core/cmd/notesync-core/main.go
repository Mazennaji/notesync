package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("notesync-core: no command provided")
		os.Exit(1)
	}
	fmt.Printf("notesync-core: received command %q (not yet implemented)\n", os.Args[1])
}
