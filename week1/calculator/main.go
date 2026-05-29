package main
import (
	
	"fmt"

	"calculator/cli"
	"calculator/evaluator"
	"calculator/parser"
)

func main(){
	input:=cli.ReadExpression()
	a,b,op,err:=parser.Parse(input)
	if err!=nil{
		fmt.Println(err)
	}
	result,err_res:=evaluator.Calculate(a,b,op)
	if err_res!=nil{
		fmt.Println(err_res)
	}
	fmt.Printf("Result: %f\n", result)
}