package main
import "fmt"

func fibo (n int) int{
	a:=0
	b:=1
	for a<n{
		next:=a+b
		a=b
		b=next
	}
	return b
}
func main(){
	fmt.Println(fibo(10))
}