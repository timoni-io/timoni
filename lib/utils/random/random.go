package random

import (
	"crypto/md5"
	"encoding/binary"
	"io"
	"math/rand"
	"strings"
	"time"

	"github.com/google/uuid"
)

const chars = "abcdefghijklmnopqrstuvwxyz0123456789"

var mdFive = md5.New()

func ShortID() string {
	return strings.ReplaceAll(
		uuid.NewString(),
		"-", "",
	)
}

func EnvID() string {
	uid := strings.ReplaceAll(
		uuid.NewString(),
		"-", "",
	)
	return "env-" + uid[:12]
}

func ID() string {
	return strings.ReplaceAll(
		uuid.New().String()+
			uuid.New().String()+
			uuid.New().String()[:16], // cut to < 90 len, becouse of backup filename length limit
		"-", "",
	)
}

// String returns random seeded string with provied length.
func String(strlen int, seed ...string) string {
	s := strings.Join(seed, "")
	if s == "" {
		s = time.Now().String()
	}

	io.WriteString(mdFive, s)

	r := rand.New(rand.NewSource(int64(binary.BigEndian.Uint64(mdFive.Sum(nil)))))

	result := make([]byte, strlen)
	for i := range result {
		result[i] = chars[r.Intn(len(chars))]
	}
	return string(result)
}
