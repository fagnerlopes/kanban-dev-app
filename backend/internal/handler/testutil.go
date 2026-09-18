package handler

import "strconv"

// itoa formats an int32 for use in URL paths in tests.
func itoa(n int32) string { return strconv.FormatInt(int64(n), 10) }
