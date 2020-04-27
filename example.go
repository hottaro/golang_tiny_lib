package main 

import (
	"github.com/hottaro/htevent"
	"github.com/hottaro/htlog"
	"fmt"
//	"./htevent"
//	"./htlog"
)

func main() {


	fmt.Println(">>>>>>>>>>> event example ...")
	// event test
	htevent.HTEvent_test()
	fmt.Println(">>>>>>>>>>> event example ... end")

	fmt.Println(">>>>>>>>>>> log example ...")
	// log test
	htlog.HTLog_test()
	fmt.Println(">>>>>>>>>>> example ... end")
}

