package bench

import (
	"testing"

	"github.com/google/uuid"
	"github.com/grishagavrin/link-shortener/internal/utils"
)

func BenchmarkFibo(b *testing.B) {
	userID := uuid.New().String()
	testShaUserID := "faa88c146770b06dbe5c90cf9f5dd17b9ac002bcdfa7a403cc59e63fb0f971b1ea9eed9817f426c3397734e43fd22172f3fbf943"

	for i := 0; i < b.N; i++ {
		_ = utils.Decode(testShaUserID, &userID)
	}
}
