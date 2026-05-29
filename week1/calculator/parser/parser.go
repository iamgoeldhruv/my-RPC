package parser

import (
	"fmt"
	"strconv"
	"strings"
)
func Parse(input string) (float64,float64,string,error){
	parts:=strings.Fields(input)
	if len(parts)!=3{
		return 0,0,"",fmt.Errorf("invalid input format, expected: <number> <operator> <number>")
	}
	a,err:=strconv.ParseFloat(parts[0],64)
	if err!=nil{
		return 0,0,"",fmt.Errorf("invalid number: %s",parts[0])

	}
	b,err:=strconv.ParseFloat(parts[2],64)
	if err!=nil{
		return 0,0,"",fmt.Errorf("invalid number: %s",parts[2])
	}
	return a,b,parts[1],nil
}