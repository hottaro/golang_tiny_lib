package htfile

import (
	"time"
    "fmt"
)

func HTFile_test() {
	file := Open("./test_cut.log")
	defer file.Close()

	file.SetFormat(CutTypeDay) 
	_, err := file.Writef("first format %v string\n", 0)
	if err != nil {
        fmt.Println("error ", err)
	}

	file.ResetFile()

	_, err = file.Writeln("file reset string")
	if err != nil {
        fmt.Println("error ", err)
	}

	nowFunc = func() time.Time {
		return time.Now().Add(time.Hour * 25)
	}

	_, err = file.Writeln("new file now")
	if err != nil {
        fmt.Println("error ", err)
	}
}

