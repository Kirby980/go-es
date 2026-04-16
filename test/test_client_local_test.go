package builder_test

import (
	"testing"

	"github.com/Kirby980/go-es/client"
)

func createTestClientLocal(t *testing.T) *client.Client {
	return createTestClient(t)
}
