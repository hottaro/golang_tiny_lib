package htevent

import (
	"fmt"
)

func HTEvent_test() {

	dispatcher := NewHTDispatcher()
	dispatcher.On("msg0", func(i int){
		fmt.Printf("msg0 dispatch ok : %d\n", i)
	})
	dispatcher.On("msg1", func(s string){
		fmt.Printf("msg1 dispatch ok : %s\n", s)
	})

	dispatcher.Handler("msg0", 0)
	err := dispatcher.Handler("msg2", 0) // error
	fmt.Println(err)
	dispatcher.Handler("msg1", "str")
	err = dispatcher.Handler("msg1", 0) // error
	fmt.Println(err)
}

