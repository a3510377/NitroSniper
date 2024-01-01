package main

type Response struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func contains[T comparable](array []T, value T) bool {
	for _, v := range array {
		if v == value {
			return true
		}
	}

	return false
}
