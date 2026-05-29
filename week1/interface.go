package main
import ("fmt"
		"math")

type Shape interface{
	Area() float64
	Perimeter() float64
}

type Circle struct{
	radius float64
}

func (c Circle) Area() float64{
	return math.Pi*c.radius*c.radius
}

func (c Circle) Perimeter() float64{
	return 2*math.Pi*c.radius
}

type Rectangle struct{
	width float64
	height float64
}

func (r Rectangle) Area() float64{
	return r.width*r.height
}

func (r Rectangle) Perimeter() float64{
	return 2*(r.width+r.height)
}

type triangle struct{
	a float64
	b float64
	c float64
}

func (t triangle) Area() float64{
	s:=(t.a+t.b+t.c)/2
	return math.Sqrt(s*(s-t.a)*(s-t.b)*(s-t.c))
}

func (t triangle) Perimeter() float64{
	return t.a+t.b+t.c
}

func main(){
	shapes:=[]Shape{
		Circle{radius:5},
		Rectangle{width:4,height:6},
		triangle{a:3,b:4,c:5},
	}
	totalArea:=0.0
	for i, shape:=range shapes{
		fmt.Printf("Shape %d: Area=%.2f, Perimeter=%.2f\n",i+1,shape.Area(),shape.Perimeter())	
		totalArea+=shape.Area()
	}
	fmt.Printf("Total Area: %.2f\n", totalArea)
}