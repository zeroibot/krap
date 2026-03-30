package conk

type Result struct {
	Success int
	Errors  map[int]error
}

type DataResult[T any] struct {
	Result
	Output map[int]T
}

func newResult() *Result {
	return &Result{
		Success: 0,
		Errors:  make(map[int]error),
	}
}

func newDataResult[T any]() *DataResult[T] {
	result := &DataResult[T]{}
	result.Errors = make(map[int]error)
	result.Output = make(map[int]T)
	return result
}

type inputData[T any] struct {
	index int
	item  T
}

type outputData[T any] struct {
	index int
	item  T
	err   error
}

type resultData struct {
	index int
	err   error
}
