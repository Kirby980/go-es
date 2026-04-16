package builder_test

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() || os.Getenv("GO_ES_INTEGRATION") != "1" {
		fmt.Fprintln(os.Stderr, "跳过集成测试：设置 GO_ES_INTEGRATION=1 以启用需要 Elasticsearch 的测试")
		os.Exit(0)
	}
	os.Exit(m.Run())
}
