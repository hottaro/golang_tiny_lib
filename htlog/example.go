package htlog

import (
	"fmt"
	"os"
)

func HTLog_test() {

	// network
	net_log := NewHTLog()
	net_log.SetHTLog("conn", `{"net":"tcp","addr":"127.0.0.1:1024"}`)
	net_log.Info("this is informational to net")

	// console
	console_log1 := NewHTLog()
	console_log1.SetHTLog("console", `{"level": "TRAC"}`)
	testConsoleCalls(console_log1)

	console_log2 := NewHTLog()
	console_log2.SetHTLog("console", `{"level":"EROR"}`)
	testConsoleCalls(console_log2)

	// file
	file_log := NewHTLog()
	file_log.SetHTLog(AdapterFile, `{"filename":"test.log",
	"level": "TRAC",
	 "rotateperm": "0666",
	"maxlines":100000,
	"maxsize":1,
	"append":true} `)

	file_log.Trace("trace")
	file_log.Debug("debug")
	file_log.Info("info")
	file_log.Warn("warning")
	file_log.Error("error")
	file_log.Emer("emergency")

	file, err := os.Stat("test.log")
	if err != nil {
		// error
	}
	if file.Mode() != 0666 {
		// error
	}

	os.Remove("test.log")

	// time format test, use default logger
	var formats = map[string]string{"ANSIC": "Mon Jan _2 15:04:05 2006",
		"UnixDate":    "Mon Jan _2 15:04:05 MST 2006",
		"RubyDate":    "Mon Jan 02 15:04:05 -0700 2006",
		"RFC822":      "02 Jan 06 15:04 MST",
		"RFC822Z":     "02 Jan 06 15:04 -0700",
		"RFC850":      "Monday, 02-Jan-06 15:04:05 MST",
		"RFC1123":     "Mon, 02 Jan 2006 15:04:05 MST",
		"RFC1123Z":    "Mon, 02 Jan 2006 15:04:05 -0700",
		"RFC3339":     "2006-01-02T15:04:05Z07:00",
		"RFC3339Nano": "2006-01-02T15:04:05.999999999Z07:00",
		"Kitchen":     "3:04PM",
		"Stamp":       "Jan _2 15:04:05",
		"StampMilli":  "Jan _2 15:04:05.000",
		"StampMicro":  "Jan _2 15:04:05.000000",
		"StampNano":   "Jan _2 15:04:05.000000000",
	}
	for timeType, format := range formats {
		file_name := fmt.Sprintf("time_log_%v", timeType)
		SetHTLog(fmt.Sprintf(`{
					"TimeFormat":"%s",
					"Console": {
						"level": "TRAC",
						"color": true
					},
					"File": {
						"filename": "%v",
						"level": "TRAC",
						"daily": true,
						"maxlines": 1000000,
						"maxsize": 1,
						"maxdays": -1,
						"append": true,
						"permit": "0660"
				}}`, format, file_name))
		fmt.Printf("========%s time format========\n", timeType)
		Trace("Trace", timeType)
		Debug("Debug", timeType)
		Info("Info", timeType)
		Warn("Warn", timeType)
		Error("Error", timeType)
		Emer("Emergency", timeType)
//		os.Remove(file_name)
	}
}

// Try each log level in decreasing order of priority.
func testConsoleCalls(bl *LocalHTLog) {
	bl.Emer("emergency")
	bl.Error("error")
	bl.Warn("warning")
	bl.Info("informational")
	bl.Debug("notice")
	bl.Trace("trace")
}

