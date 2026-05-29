package main
import (

	// fmt package is used for printing output
	"fmt"

	// os package is used for file handling
	"os"

	// strings package helps manipulate strings
	"strings"

	// unicode package helps check characters
	"unicode"

	// sort package helps sort slices
	"sort"
)
type WordFrequency struct{
	word string
	count int
}
func main(){
	data, err := os.ReadFile("sample.txt")
	if err!=nil{
		fmt.Println("Error reading file:",err)
		return
	}
	text := string(data)
	lowerText:=strings.ToLower(text)
	cleanText:=strings.Map(func(r rune) rune{
		if unicode.IsLetter(r) || unicode.IsSpace(r)|| unicode.IsDigit(r){
			return r
		}
		return -1

	},lowerText)
	words:=strings.Fields(cleanText)
	wordCount:=make(map[string]int)
	for _, word:=range words{
		wordCount[word]++
	}
	var frequencies []WordFrequency
	for word, count := range wordCount{
		item:=WordFrequency{word:word,count:count}
		frequencies=append(frequencies,item)
	}
	sort.Slice(frequencies, func(i, j int) bool {

		// Return true if i should come before j

		return frequencies[i].count > frequencies[j].count
	})
	fmt.Println("Top 10 Most Common Words")
	fmt.Println("-------------------------")

	// Decide limit

	limit := 10

	// If fewer than 10 words exist

	if len(frequencies) < 10 {

		limit = len(frequencies)
	}

	///////////////////////////////////////////////////////////
	// STEP 12: PRINT RESULTS
	///////////////////////////////////////////////////////////

	for i := 0; i < limit; i++ {

		fmt.Printf(

			"%d. %s -> %d times\n",

			i+1,

			frequencies[i].word,

			frequencies[i].count,
		)
	}
}
