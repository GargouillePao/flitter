package main

import (
	"fmt"
)

func main() {
	strs := make([]string, 0)
	for i := 0; i < 10; i++ {
		strs = append(strs, fmt.Sprintf("Hello%d", i))
	}
	fmt.Println(strs)

	for i := 0; i < len(strs); i++ {
		index, str := i, strs[i]
		if str == "Hello2" || str == "Hello9" {
			fmt.Println(str, len(strs), index)
			if index >= len(strs)-1 {
				strs = strs[:index]
				fmt.Println(len(strs))
			} else {
				strs = append(strs[:index], strs[index+1:]...)
			}

		}
	}
	//for index, str := range strs {

	//}

	fmt.Println(strs)
}
