package main
import (
	"fmt"
	"os"
	"path/filepath"
)

func printTree(path string,prefix string) error{
	files,err:=os.ReadDir(path)
	if err!=nil{
		return err
	}
	for i,file:=range files{
		isLast:=i==len(files)-1
		var connector string
		if isLast{
			connector="└── "
		}else{
			connector="├── "
		}
		fmt.Println(prefix+connector+file.Name())
		if file.IsDir(){
			var newPrefix string
			if isLast{
				newPrefix=prefix+"    "
			}else{
				newPrefix=prefix+"│   "
			}
			fullPath:=filepath.Join(path,file.Name())
			err:=printTree(fullPath,newPrefix)
			if err!=nil{
				return err
			}

		}
	}
	return nil;

}

func main(){
	if len(os.Args)>2{
		fmt.Println("Usage: tree [directory]")
		return
	}

	target_dir:=os.Args[1]

	dir_info,error:=os.Stat(target_dir)
	if error!=nil{
		fmt.Printf("Error: %s\n", error)
		return
	}
	if !dir_info.IsDir(){
		fmt.Printf("Error: %s is not a directory\n", target_dir)
		return
	}
	fmt.Println(filepath.Base(target_dir))
	err:=printTree(target_dir,"")
	if err!=nil{
		fmt.Printf("Error: %s\n", err)
	}
	return

}