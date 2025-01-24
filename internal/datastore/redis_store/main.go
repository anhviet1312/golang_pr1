package redis_store

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func dbKeySendOtpCode(userId int64) string {
	return fmt.Sprintf("otp-code:%d", userId)
}

func GetOtpCode(ctx context.Context, cmd redis.Cmdable, userId int64) (string, error) {
	otpCode, err := cmd.Get(ctx, dbKeySendOtpCode(userId)).Result()
	if err != nil {
		return "", err
	}

	return otpCode, nil
}

func SetOtpCode(ctx context.Context, cmd redis.Cmdable, userId int64, otpCode string) (bool, error) {
	err := cmd.Set(ctx, dbKeySendOtpCode(userId), otpCode, 5*time.Minute).Err()
	if err != nil {
		return false, err
	}

	return true, nil
}
