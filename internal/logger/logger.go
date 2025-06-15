package logger

import (
	"fmt"
	"log"
)

func Debug(format string, v ...any) {
	log.Printf(fmt.Sprintf("[DEBUG] %s\n", format), v...)
}

func Info(format string, v ...any) {
	log.Printf(fmt.Sprintf("[INFO] %s\n", format), v...)
}

func Error(format string, v ...any) {
	log.Printf(fmt.Sprintf("[ERROR] %s\n", format), v...)
}

func Fatal(format string, v ...any) {
	log.Fatalf(fmt.Sprintf("[FATAL] %s\n", format), v...)
}
