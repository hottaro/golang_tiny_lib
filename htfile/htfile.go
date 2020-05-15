package htfile

import (
	"os"
	"sync"
	"time"
	"fmt"
	"strings"
)

type TCut string

const (
	CutTypeHour  TCut = "06010215"
	CutTypeDay   TCut = "060102"
	CutTypeMonth TCut = "0601"

	LogFlag = os.O_WRONLY | os.O_CREATE | os.O_APPEND

	LogMode = 0666
)

type HTFile struct {
	file        *os.File
	cutType  string
	orgFilename string

	flag int
	mode os.FileMode

	filenameMu   sync.Mutex
	destFilename string
	destKey      string 
}

func Open(filename string) *HTFile {
	return &HTFile{
		flag: LogFlag,
		mode: LogMode,

		orgFilename: filename,
		cutType:  string(CutTypeHour), // 默认按小时
	}
}

var nowFunc = time.Now

func (f *HTFile) SetFormat(cutType TCut) {
	f.cutType = string(cutType)
}

func (f *HTFile) SetFlag(flag int) {
	f.flag = flag
}

func (f *HTFile) SetMode(mode os.FileMode) {
	f.mode = mode
}

func (f *HTFile) ResetFile() error {
	now := nowFunc()
	f.filenameMu.Lock()
	defer f.filenameMu.Unlock()

	key := now.Format(f.cutType)
	name := f.orgFilename + "." + key

	file, err := os.OpenFile(name, f.flag, f.mode)
	if err != nil {
		f.filenameMu.Unlock()
		return err
	}

	if f.file != nil {
		f.file.Close()
	}

	f.file, f.destFilename, f.destKey = file, name, key
	return nil
}

func (f *HTFile) Writeb(b []byte) (n int, err error) {
	now := nowFunc()
	f.filenameMu.Lock()
	key := now.Format(f.cutType)
	if key != f.destKey {
		if f.file != nil {
			f.file.Close()
		}

		name := f.orgFilename + "." + key

		f.file, err = os.OpenFile(name, f.flag, f.mode)
		if err != nil {
			f.filenameMu.Unlock()
			return n, err
		}

		f.destFilename, f.destKey = name, key
	}
	f.filenameMu.Unlock()

	return f.file.Write(b)
}

func (f *HTFile) Write(s string) (n int, err error) {
	return f.Writeb([]byte(s))
}

func (f *HTFile) Writeln(s string) (n int, err error) {
	return f.Write(s + "\n")
}

func (f *HTFile) Writef(format interface{}, v ...interface{}) (n int, err error) {
	return f.Write(formatLog(format, v...))
}

func (f *HTFile) Close() error {
	return f.file.Close()
}

func formatLog(f interface{}, v ...interface{}) string {
	var msg string
	switch f.(type) {
	case string:
		msg = f.(string)
		if len(v) == 0 {
			return msg
		}
		if strings.Contains(msg, "%") && !strings.Contains(msg, "%%") {
			//format string
		} else {
			//do not contain format char
			msg += strings.Repeat(" %v", len(v))
		}
	default:
		msg = fmt.Sprint(f)
		if len(v) == 0 {
			return msg
		}
		msg += strings.Repeat(" %v", len(v))
	}
	return fmt.Sprintf(msg, v...)
}
