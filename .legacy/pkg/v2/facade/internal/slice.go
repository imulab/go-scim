package internal

type Slice []interface{}

func (s Slice) StringTyped() []string {
	var arr []string
	for _, each := range s {
		arr = append(arr, each.(string))
	}
	return arr
}

func (s Slice) Int64Typed() []int64 {
	var arr []int64
	for _, each := range s {
		arr = append(arr, each.(int64))
	}
	return arr
}

func (s Slice) Float64Typed() []float64 {
	var arr []float64
	for _, each := range s {
		arr = append(arr, each.(float64))
	}
	return arr
}

func (s Slice) BoolTyped() []bool {
	var arr []bool
	for _, each := range s {
		arr = append(arr, each.(bool))
	}
	return arr
}
