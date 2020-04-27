package htlog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// 默认日志输出
var defaultHTLog *LocalHTLog

// 日志等级，从0-7，日优先级由高到低
const (
	LevelEmergency     = iota // 系统级紧急，比如磁盘出错，内存异常，网络不可用等
	LevelError                // 用户级错误
	LevelWarning              // 用户级警告
	LevelInformational        // 用户级信息
	LevelDebug                // 用户级调试
	LevelTrace                // 用户级基本输出
)

// 日志等级和描述映射关系
var LevelMap = map[string]int{
	"EMER": LevelEmergency,
	"EROR": LevelError,
	"WARN": LevelWarning,
	"INFO": LevelInformational,
	"DEBG": LevelDebug,
	"TRAC": LevelTrace,
}

// 注册实现的适配器， 当前支持控制台，文件和网络输出
var adapters = make(map[string]HTLog)

// 日志记录等级字段
var levelPrefix = [LevelTrace + 1]string{
	"EMER",
	"EROR",
	"WARN",
	"INFO",
	"DEBG",
	"TRAC",
}

const (
	logTimeDefaultFormat = "2006-01-02 15:04:05" // 日志输出默认格式
	AdapterConsole       = "console"             // 控制台输出配置项
	AdapterFile          = "file"                // 文件输出配置项
	AdapterConn          = "conn"                // 网络输出配置项
)

// log provider interface
type HTLog interface {
	Init(config string) error
	LogWrite(when time.Time, msg interface{}, level int) error
	Destroy()
}

// 日志输出适配器注册，log需要实现Init，LogWrite，Destroy方法
func Register(name string, log HTLog) {
	if log == nil {
		panic("logs: Register provide is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("logs: Register called twice for provider " + name)
	}
	adapters[name] = log
}

type loginfo struct {
	Time    string
	Level   string
	Path    string
	Name    string
	Content string
}

type nameHTLog struct {
	HTLog
	name   string
	config string
}

type LocalHTLog struct {
	lock       sync.Mutex
	init       bool
	outputs    []*nameHTLog
	appName    string
	callDepth  int
	timeFormat string
	usePath    string
}

func NewHTLog(depth ...int) *LocalHTLog {
	dep := append(depth, 2)[0]
	l := new(LocalHTLog)
	// appName用于记录网络传输时标记的程序发送方，
	// 通过环境变量APPSN进行设置,默认为NONE,此时无法通过网络日志检索区分不同服务发送方
	appSn := os.Getenv("APPSN")
	if appSn == "" {
		appSn = "NONE"
	}
	l.appName = "[" + appSn + "]"
	l.callDepth = dep
	l.SetHTLog(AdapterConsole)
	l.timeFormat = logTimeDefaultFormat
	return l
}

//配置文件
type logConfig struct {
	TimeFormat string         `json:"TimeFormat"`
	Console    *consoleHTLog `json:"Console,omitempty"`
	File       *fileHTLog    `json:"File,omitempty"`
	Conn       *connHTLog    `json:"Conn,omitempty"`
}

func init() {
	defaultHTLog = NewHTLog(3)
}

func (this *LocalHTLog) SetHTLog(adapterName string, configs ...string) error {
	this.lock.Lock()
	defer this.lock.Unlock()

	if !this.init {
		this.outputs = []*nameHTLog{}
		this.init = true
	}

	config := append(configs, "{}")[0]
	var num int = -1
	var i int
	var l *nameHTLog
	for i, l = range this.outputs {
		if l.name == adapterName {
			if l.config == config {
				//配置没有变动，不重新设置
				return fmt.Errorf("you have set same config for this adaptername %s", adapterName)
			}
			l.HTLog.Destroy()
			num = i
			break
		}
	}
	htlog, ok := adapters[adapterName]
	if !ok {
		return fmt.Errorf("unknown adaptername %s (forgotten Register?)", adapterName)
	}

	err := htlog.Init(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "htlog Init <%s> err:%v, %s output ignore!\n",
			adapterName, err, adapterName)
		return err
	}
	if num >= 0 {
		this.outputs[i] = &nameHTLog{name: adapterName, HTLog: htlog, config: config}
		return nil
	}
	this.outputs = append(this.outputs, &nameHTLog{name: adapterName, HTLog: htlog, config: config})
	return nil
}

func (this *LocalHTLog) DelHTLog(adapterName string) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	outputs := []*nameHTLog{}
	for _, lg := range this.outputs {
		if lg.name == adapterName {
			lg.Destroy()
		} else {
			outputs = append(outputs, lg)
		}
	}
	if len(outputs) == len(this.outputs) {
		return fmt.Errorf("logs: unknown adaptername %s (forgotten Register?)", adapterName)
	}
	this.outputs = outputs
	return nil
}

// 设置日志起始路径
func (this *LocalHTLog) SetLogPathTrim(trimPath string) {
	this.usePath = trimPath
}

func (this *LocalHTLog) writeToHTLogs(when time.Time, msg *loginfo, level int) {
	for _, l := range this.outputs {
		if l.name == AdapterConn {
			//网络日志，使用json格式发送,此处使用结构体，用于类似ElasticSearch功能检索
			err := l.LogWrite(when, msg, level)
			if err != nil {
				fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", l.name, err)
			}
			continue
		}

		msgStr := when.Format(this.timeFormat) + " [" + msg.Level + "] " + "[" + msg.Path + "] " + msg.Content
		err := l.LogWrite(when, msgStr, level)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", l.name, err)
		}
	}
}

func (this *LocalHTLog) writeMsg(logLevel int, msg string, v ...interface{}) error {
	if !this.init {
		this.SetHTLog(AdapterConsole)
	}
	msgSt := new(loginfo)
	src := ""
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	when := time.Now()
	_, file, lineno, ok := runtime.Caller(this.callDepth)
	var strim string = "src/"
	if this.usePath != "" {
		strim = this.usePath
	}
	if ok {

		src = strings.Replace(
			fmt.Sprintf("%s:%d", stringTrim(file, strim), lineno), "%2e", ".", -1)
	}

	msgSt.Level = levelPrefix[logLevel]
	msgSt.Path = src
	msgSt.Content = msg
	msgSt.Name = this.appName
	msgSt.Time = when.Format(this.timeFormat)
	this.writeToHTLogs(when, msgSt, logLevel)

	return nil
}

func (this *LocalHTLog) Fatal(format string, args ...interface{}) {
	this.Emer("###Exec Panic:"+format, args...)
	os.Exit(1)
}

func (this *LocalHTLog) Panic(format string, args ...interface{}) {
	this.Emer("###Exec Panic:"+format, args...)
	panic(fmt.Sprintf(format, args...))
}

// Emer Log EMERGENCY level message.
func (this *LocalHTLog) Emer(format string, v ...interface{}) {
	this.writeMsg(LevelEmergency, format, v...)
}

// Error Log ERROR level message.
func (this *LocalHTLog) Error(format string, v ...interface{}) {
	this.writeMsg(LevelError, format, v...)
}

// Warn Log WARNING level message.
func (this *LocalHTLog) Warn(format string, v ...interface{}) {
	this.writeMsg(LevelWarning, format, v...)
}

// Info Log INFO level message.
func (this *LocalHTLog) Info(format string, v ...interface{}) {
	this.writeMsg(LevelInformational, format, v...)
}

// Debug Log DEBUG level message.
func (this *LocalHTLog) Debug(format string, v ...interface{}) {
	this.writeMsg(LevelDebug, format, v...)
}

// Trace Log TRAC level message.
func (this *LocalHTLog) Trace(format string, v ...interface{}) {
	this.writeMsg(LevelTrace, format, v...)
}

func (this *LocalHTLog) Close() {

	for _, l := range this.outputs {
		l.Destroy()
	}
	this.outputs = nil

}

func (this *LocalHTLog) Reset() {
	for _, l := range this.outputs {
		l.Destroy()
	}
	this.outputs = nil
}

func (this *LocalHTLog) SetCallDepth(depth int) {
	this.callDepth = depth
}

// GetlocalHTLog returns the defaultHTLog
func GetlocalHTLog() *LocalHTLog {
	return defaultHTLog
}

// Reset will remove all the adapter
func Reset() {
	defaultHTLog.Reset()
}

func SetLogPathTrim(trimPath string) {
	defaultHTLog.SetLogPathTrim(trimPath)
}

// param 可以是log配置文件名，也可以是log配置内容,默认DEBUG输出到控制台
func SetHTLog(param ...string) error {
	if 0 == len(param) {
		//默认只输出到控制台
		defaultHTLog.SetHTLog(AdapterConsole)
		return nil
	}

	c := param[0]
	conf := new(logConfig)
	err := json.Unmarshal([]byte(c), conf)
	if err != nil { //不是json，就认为是配置文件，如果都不是，打印日志，然后退出
		// Open the configuration file
		fd, err := os.Open(c)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open %s for configure: %s\n", c, err)
			os.Exit(1)
			return err
		}

		contents, err := ioutil.ReadAll(fd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not read %s: %s\n", c, err)
			os.Exit(1)
			return err
		}
		err = json.Unmarshal(contents, conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not Unmarshal %s: %s\n", contents, err)
			os.Exit(1)
			return err
		}
	}
	if conf.TimeFormat != "" {
		defaultHTLog.timeFormat = conf.TimeFormat
	}
	if conf.Console != nil {
		console, _ := json.Marshal(conf.Console)
		defaultHTLog.SetHTLog(AdapterConsole, string(console))
	}
	if conf.File != nil {
		file, _ := json.Marshal(conf.File)
		fmt.Println("", string(file))
		defaultHTLog.SetHTLog(AdapterFile, string(file))
	}
	if conf.Conn != nil {
		conn, _ := json.Marshal(conf.Conn)
		defaultHTLog.SetHTLog(AdapterConn, string(conn))
	}
	return nil
}

// Painc logs a message at emergency level and panic.
func Painc(f interface{}, v ...interface{}) {
	defaultHTLog.Panic(formatLog(f, v...))
}

// Fatal logs a message at emergency level and exit.
func Fatal(f interface{}, v ...interface{}) {
	defaultHTLog.Fatal(formatLog(f, v...))
}

// Emer logs a message at emergency level.
func Emer(f interface{}, v ...interface{}) {
	defaultHTLog.Emer(formatLog(f, v...))
}

// Error logs a message at error level.
func Error(f interface{}, v ...interface{}) {
	defaultHTLog.Error(formatLog(f, v...))
}

// Warn logs a message at warning level.
func Warn(f interface{}, v ...interface{}) {
	defaultHTLog.Warn(formatLog(f, v...))
}

// Info logs a message at info level.
func Info(f interface{}, v ...interface{}) {
	defaultHTLog.Info(formatLog(f, v...))
}

// Notice logs a message at debug level.
func Debug(f interface{}, v ...interface{}) {
	defaultHTLog.Debug(formatLog(f, v...))
}

// Trace logs a message at trace level.
func Trace(f interface{}, v ...interface{}) {
	defaultHTLog.Trace(formatLog(f, v...))
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

func stringTrim(s string, cut string) string {
	ss := strings.SplitN(s, cut, 2)
	if 1 == len(ss) {
		return ss[0]
	}
	return ss[1]
}
