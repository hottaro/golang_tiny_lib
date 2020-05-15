package main 

import (
	"github.com/hottaro/golang_tiny_lib/htevent"
	"github.com/hottaro/golang_tiny_lib/htlog"
	"github.com/hottaro/golang_tiny_lib/htfile"
	"fmt"
//	"./htevent"
//	"./htlog"
// "./htfile"
)

func main() {

	// event test
	fmt.Println(">>>>>>>>>>> event example ...")
	htevent.HTEvent_test()
	fmt.Println(">>>>>>>>>>> event example ... end")

	// log test
	fmt.Println(">>>>>>>>>>> log example ...")
	htlog.HTLog_test()
	fmt.Println(">>>>>>>>>>> example ... end")

	// file test
	fmt.Println(">>>>>>>>>>> log example ...")
    htfile.HTFile_test()	
	fmt.Println(">>>>>>>>>>> example ... end")
}

