package util

import (
	"fmt"
	"github.com/LuisZhou/lpge/conf"
	"github.com/LuisZhou/lpge/log"
	"runtime"
	"unicode"
)

func addSegment(inrune, segment []rune) []rune {
	if len(segment) == 0 {
		return inrune
	}
	if len(inrune) != 0 {
		inrune = append(inrune, '_')
	}
	inrune = append(inrune, segment...)
	return inrune
}

// CamelCaseToUnderscore converts from camel case form to underscore separated form.
// Ex.: MyFunc => my_func
func CamelCaseToUnderscore(str string) string {
	var output []rune
	var segment []rune
	for _, r := range str {

		// not treat number as separate segment
		if !unicode.IsLower(r) && string(r) != "_" && !unicode.IsNumber(r) {
			output = addSegment(output, segment)
			segment = nil
		}
		segment = append(segment, unicode.ToLower(r))
	}
	output = addSegment(output, segment)
	return string(output)
}

func recoverAndLog() (err error) {
	if r := recover(); r != nil {
		if conf.LenStackBuf > 0 {
			buf := make([]byte, conf.LenStackBuf)
			l := runtime.Stack(buf, false)
			err = fmt.Errorf("%v: %s", r, buf[:l])
		} else {
			err = fmt.Errorf("%v", r)
		}
		log.Error(err.Error())
	}
	return
}
