package advisor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestPublicHidesOperatorDetail(t *testing.T) {
	apiErr := errors.New("OpenAI API 401: Incorrect API key provided: sk-abc")
	auth := shopperErr("顧問暫時無法使用，請稍後再試", fmt.Errorf("請確認 OPENAI_API_KEY: %w", apiErr))
	cases := []struct {
		err  error
		want string
	}{
		{auth, "顧問暫時無法使用，請稍後再試"},
		{fmt.Errorf("wrapped: %w", auth), "顧問暫時無法使用，請稍後再試"},
		{errors.New(`Post "https://api.openai.com/v1/chat/completions": dial tcp: i/o timeout`), "顧問暫時無法回覆，請稍後再試"},
		{fmt.Errorf("stream: %w", context.DeadlineExceeded), "回覆時間過長，請把問題縮小一些再試一次"},
	}
	for _, c := range cases {
		if got := Public(c.err); got != c.want {
			t.Errorf("Public(%q) = %q, want %q", c.err, got, c.want)
		}
	}
	// The log side keeps the cause.
	if !strings.Contains(auth.Error(), "sk-abc") || !errors.Is(auth, apiErr) {
		t.Errorf("Error() lost the cause: %q", auth.Error())
	}
}
