package htlog

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

type brush func(string) string

func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

/*
前景色   背景色  
30  	40	  黑色
31  	41	  红色
32  	42	  绿色
33  	43    黄色
34  	44    蓝色
35  	45 	  紫色
36  	46 	  青色
37  	47	  白色 
*/
var colors = []brush{
	newBrush("1;41"), // Emergency        
	newBrush("1;31"), // Error           
	newBrush("0;33"), // Warn            
	newBrush("0;37"), // Informational  
	newBrush("0;32"), // Debug         
	newBrush("0;36"), // Trace        
	//newBrush("1;35"), // Alert     
	//newBrush("1;34"), // Critical 
}

type consoleHTLog struct {
	sync.Mutex
	Level    string `json:"level"`
	Colorful bool   `json:"color"`
	LogLevel int
}

func (c *consoleHTLog) Init(jsonConfig string) error {
	if len(jsonConfig) == 0 {
		return nil
	}
	if jsonConfig != "{}" {
		fmt.Fprintf(os.Stdout, "consoleHTLog Init:%s\n", jsonConfig)
	}

	err := json.Unmarshal([]byte(jsonConfig), c)
	if runtime.GOOS == "windows" {
		c.Colorful = false
	}

	if l, ok := LevelMap[c.Level]; ok {
		c.LogLevel = l
		return nil
	}

	return err
}

func (c *consoleHTLog) LogWrite(when time.Time, msgText interface{}, level int) error {
	if level > c.LogLevel {
		return nil
	}
	msg, ok := msgText.(string)
	if !ok {
		return nil
	}
	if c.Colorful {
		msg = colors[level](msg)
	}
	c.printlnConsole(when, msg)
	return nil
}

func (c *consoleHTLog) Destroy() {

}

func (c *consoleHTLog) printlnConsole(when time.Time, msg string) {
	c.Lock()
	defer c.Unlock()
	os.Stdout.Write(append([]byte(msg), '\n'))
}

func init() {
	Register(AdapterConsole, &consoleHTLog{
		LogLevel: LevelDebug,
		Colorful: runtime.GOOS != "windows",
	})
}
