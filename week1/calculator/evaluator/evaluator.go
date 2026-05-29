package evaluator
import("fmt" 
"errors"
"math")

var ErrorDivisionByZero=errors.New("division by zero is not allowed")
var unsupportedOperationError=errors.New("unsupported operation")

type NegativeSqrtError struct{
	Num float64
}
func (e NegativeSqrtError) Error() string{
	return fmt.Sprintf("cannot compute the sqrt root of negative number: %f",e.Num)
}

func Calculate(a,b float64, operator string) (float64,error){
	switch operator {

	case "+":
		return a + b, nil

	case "-":
		return a - b, nil

	case "*":
		return a * b, nil

	case "/":
		if b == 0 {
			return 0,
				fmt.Errorf(
					"division by zero: %w",
					ErrorDivisionByZero,
				)
		}
		return a / b, nil

	default:
		return 0,
			fmt.Errorf(
				"unsupported operation %s: %w",
				operator,
				unsupportedOperationError,
			)
	}
}

func Sqrt(number float64) (float64,error){
	if number<0{
		return 0,NegativeSqrtError{Num:number}

	}
	return math.Sqrt(number), nil
}