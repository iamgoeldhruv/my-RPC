package main
import("fmt" 
"errors"
"math")

var ErrorDivisionByZero=errors.New("division by zero is not allowed")
var UnsupportedOperatioNError=errors.New("unsupported operation")

type NegativeSqrtError struct{
	Number float64
}
func (e NegativeSqrtError) Error() string{
	return fmt.Sprintf("cannot compute square root of negative number: %f",e.Number)	
}
func calculate(a,b float64, operator string) (float64,error){
	switch operator{
	case "+":
		return a+b, nil
	case "-":
		return a-b,nil
	case "*":
		return a*b,nil
	case "/":
		if b==0{
			return 0, fmt.Errorf("division by zero: %w",ErrorDivisionByZero)
		}
		return a/b,nil
	default:
		return 0, fmt.Errorf("unsupported operation: %s: %w",operator,UnsupportedOperatioNError)


	}

}
func sqrt(number float64)(float64,error){
	if number<0{
		return 0,NegativeSqrtError{Number:number}

	}
	return math.Sqrt(number),nil
}
func main(){
	add,err:=calculate(10,5,"+")
	if err!=nil{
		fmt.Println("Error:",err)

	}else{
		fmt.Printf("10 + 5 = %.2f\n",add)
	}
	sub,err:=calculate(10,5,"-")
	if err!=nil{
		fmt.Println("Error:",err)
	}else{
		fmt.Printf("10 - 5 = %.2f\n",sub)
	}

	divByZero,err:=calculate(10,0,"/")
	if err!=nil{
		fmt.Println("Error:",err)
		if errors.Is(err,ErrorDivisionByZero){
			fmt.Println("Please provide a non-zero divisor.")
		}
	}else{
		fmt.Printf("10 / 0 = %.2f\n",divByZero)
	}

	tests,err:=calculate(10,5,"^")
	if err!=nil{
		fmt.Println("Error:",err)
		if errors.Is(err,UnsupportedOperatioNError){
			fmt.Println("Please provide a supported operator.")
		}
	}else{
		fmt.Printf("10 ^ 5 = %.2f\n",tests)
	}
}