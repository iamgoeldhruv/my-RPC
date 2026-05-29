package main
import (
	"fmt"
	"math/rand"
	"time"
)

type Result struct{
	ID int
	Duration time.Duration

}
func worker(id int ,ch chan Result) {
	sleepTime :=
		time.Duration(
			rand.Intn(3000),
		) * time.Millisecond
	time.Sleep(sleepTime)
	ch <- Result{ID:id, Duration:sleepTime}
}

func main(){
	rand.Seed(time.Now().UnixNano())
	ch:=make(chan Result)
	for i:=1;i<=10;i++{
		go worker(i,ch)
	}
	fmt.Println("Results in comparison order")
	for i:=1;i<=10;i++{
		result1:=<-ch
		fmt.Printf("Worker %d completed in %v\n",result1.ID,result1.Duration)
	}

}