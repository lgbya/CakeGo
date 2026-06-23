package client

func Rand(min int, max int) int {
	return r.Intn(max-min+1) + min
}
