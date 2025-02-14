package pkg

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/hiendaovinh/toolkit/pkg/db"
	"github.com/mozillazg/go-unidecode"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"golang.org/x/text/unicode/norm"
	"io"
	"math/big"
	math_rand "math/rand"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	ADDRESS_SMTP string = "smtp.yandex.com:587"
)

type Email struct {
	From    string
	To      []string
	Subject string
	Body    string
}

func GenerateRandomID() int64 {
	return math_rand.Int63()
}

func GenerateOTP() (string, error) {
	otp := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		otp += fmt.Sprintf("%d", num.Int64())
	}
	return otp, nil
}

func SendMail(email *Email, auth sasl.Client) error {
	// Set up header
	headers := make(map[string]string)
	headers["From"] = email.From
	headers["To"] = strings.Join(email.To, ",")
	headers["Subject"] = email.Subject
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + email.Body)

	err := smtp.SendMail(ADDRESS_SMTP, auth, email.From, email.To, strings.NewReader(msg.String()))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	return nil
}

func FetchContent(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func GetDb() (*bun.DB, error) {
	sqldb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(os.Getenv("DB_DSN")),
	))

	db := bun.NewDB(sqldb, pgdialect.New())
	return db, nil
}

func GetRedis() (redis.UniversalClient, error) {
	var dbRedis redis.UniversalClient
	var err error

	clusterRedisQuestionnaire := os.Getenv("CLUSTER_REDIS")
	if clusterRedisQuestionnaire != "" {
		clusterOpts, err := redis.ParseClusterURL(clusterRedisQuestionnaire)
		if err != nil {
			return nil, err
		}
		dbRedis = redis.NewClusterClient(clusterOpts)
	} else {
		dbRedis, err = db.InitRedis(&db.RedisConfig{
			URL: os.Getenv("REDIS_DB"),
		})
		if err != nil {
			return nil, err
		}
	}
	return dbRedis, nil
}

func NormalizeURL(inputURL string) (string, error) {
	decoded, err := url.QueryUnescape(inputURL)
	if err != nil {
		return "", err
	}

	normalized := norm.NFC.String(decoded)
	noAccent := unidecode.Unidecode(normalized)
	return noAccent, nil
}

func GetStoryID(content string) int {
	regex, err := regexp.Compile("book_detail\" content=\"(\\d*?)\"")
	if err != nil {
		return 0
	}
	storyIDString := regex.FindStringSubmatch(content)
	if len(storyIDString) != 2 {
		return 0
	}

	num, err := strconv.Atoi(storyIDString[1])
	if err != nil {
		fmt.Println("Lỗi khi chuyển đổi:", err)
		return 0
	}
	return num
}
