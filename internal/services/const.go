package services

import (
	"fmt"
	"time"
)

const (
	CacheTtl5Mins = 5 * time.Minute
)

func DBKeyUserByUsername(username string) string {
	return fmt.Sprintf("user:%s", username)
}
