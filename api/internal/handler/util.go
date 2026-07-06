package handler

import "strconv"

func parseID(raw string) (int64, error) {
	return strconv.ParseInt(raw, 10, 64)
}
