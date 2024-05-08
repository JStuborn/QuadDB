// ./util/logger.go

package util

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	infoColor  = color.New(color.Bold, color.BgBlue, color.FgWhite)
	warnColor  = color.New(color.Bold, color.BgYellow, color.FgWhite)
	errorColor = color.New(color.Bold, color.BgRed, color.FgWhite)
	debugColor = color.New(color.Bold, color.BgCyan, color.FgWhite)
)

func Info(format string, a ...interface{}) {
	infoColor.Print("[ INFO  ]")
	infoColor.DisableColor()
	fmt.Printf(" "+format+" \n", a...)
	infoColor.EnableColor()
}

func Warn(format string, a ...interface{}) {
	warnColor.Print("[ WARN  ]")
	warnColor.DisableColor()
	fmt.Printf(" "+format+" \n", a...)
	warnColor.EnableColor()
}

func Error(format string, a ...interface{}) {
	errorColor.Print("[ ERROR ]")
	errorColor.DisableColor()
	fmt.Printf(" "+format+" \n", a...)
	errorColor.EnableColor()
}

func Debug(format string, a ...interface{}) {
	debugColor.Print("[ DEBUG ]")
	debugColor.DisableColor()
	fmt.Printf(" "+format+" \n", a...)
	debugColor.EnableColor()
}
