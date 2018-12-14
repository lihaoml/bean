package util

import "reflect"

func Contains(a interface{}, e interface{}) bool {
	v := reflect.ValueOf(a)
	for i := 0; i < v.Len(); i++ {
		if v.Index(i).Interface() == e {
			return true
		}
	}
	return false
}

func AverageSlice(sl []float64) float64 {
	var i int
	var v float64
	vs := 0.0
	for i, v = range sl {
		vs = vs + v
	}
	return vs / float64(i+1)
}

func MinOf(vars ...int) int {
	min := vars[0]
	for _, i := range vars {
		if min > i {
			min = i
		}
	}
	return min
}

func MaxOf(vars ...int) int {
	max := vars[0]
	for _, i := range vars {
		if max < i {
			max = i
		}
	}
	return max
}