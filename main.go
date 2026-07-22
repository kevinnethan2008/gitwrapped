package main 

import (
	"fmt"
	"os/exec"

)

func main() {

	cmd := exec.Commnad("git", "log", "--pretty=format:%H|%an|%ad|%s")

	output, err := cmd.Output()
	if err != nil {
	fmt.Println("Error running git log:",err)
	return
	}

	fmt.Println(string(output))


}

