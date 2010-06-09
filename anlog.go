

package anlog



import (
	"fmt"
	"os"
)


func Info(format string, v ...interface{}){fmt.Printf("[info] " + format, v);}
func Warn(format string, v ...interface{}) {fmt.Printf("[warning] " + format, v);}
func Error(format string, v ...interface{}) {fmt.Fprintf(os.Stderr,"[error] " + format, v);}

