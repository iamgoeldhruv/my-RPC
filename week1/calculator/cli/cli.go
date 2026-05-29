package cli
import (
	"bufio"
	"fmt"
	"os"
)

func ReadExpression()string{
	fmt.Print("Enter an expression (e.g., 3 + 4): ")
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return text
}