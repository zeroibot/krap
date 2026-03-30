package sys

import (
	"fmt"

	"github.com/zeroibot/fn/ds"
	"github.com/zeroibot/fn/str"
	"github.com/zeroibot/rdb/ze"
)

// Prints request logs and error; prints OK if no error
func DisplayOutput(rq *ze.Request, err error) {
	if rq != nil {
		displayOutput(rq)
	}
	if err == nil {
		fmt.Println(okMessage)
	} else {
		DisplayError(err)
	}
}

// Prints request logs and error
func DisplayResult(rq *ze.Request, err error) {
	if rq != nil {
		displayOutput(rq)
	}
	if err != nil {
		DisplayError(err)
	}
}

// Prints error
func DisplayError(err error) {
	if err != nil {
		fmt.Println("Error:", err)
	}
}

// Prints data, request logs, and error
func DisplayData[T any](data *T, rq *ze.Request, err error) {
	if rq != nil {
		displayOutput(rq)
	}
	if err == nil {
		if data == nil {
			fmt.Println("Data: null")
		} else {
			output, err := str.IndentedJSON(data)
			if err == nil {
				fmt.Println(output)
			} else {
				DisplayError(err)
			}
		}
	} else {
		DisplayError(err)
	}
}

// Prints list items, request logs, and error
func DisplayList[T any](list *ds.List[*T], rq *ze.Request, err error) {
	if rq != nil {
		displayOutput(rq)
	}
	if err == nil {
		for i, item := range list.Items {
			if item == nil {
				fmt.Printf("%d: nil\n", i+1)
			} else {
				fmt.Printf("%d: %v\n", i+1, *item)
			}
		}
		fmt.Println("Count:", list.Count)
	} else {
		DisplayError(err)
	}
}

func displayOutput(rq *ze.Request) {
	output := rq.Output()
	if output != "" {
		fmt.Println(output)
	}
}
