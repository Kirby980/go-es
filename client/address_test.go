package client

import (
	"testing"

	"github.com/Kirby980/go-es/config"
	"github.com/Kirby980/go-es/logger"
)

func TestGetAddress_FirstCallReturnsFirstAddress(t *testing.T) {
	c := &Client{
		config:    &config.Config{},
		addresses: []string{"http://node1:9200", "http://node2:9200", "http://node3:9200"},
		logger:    logger.NopLogger{},
	}

	first := c.GetAddress()
	if first != "http://node1:9200" {
		t.Errorf("first call should return addresses[0], got %q", first)
	}
}

func TestGetAddress_RoundRobin(t *testing.T) {
	addrs := []string{"http://node1:9200", "http://node2:9200", "http://node3:9200"}
	c := &Client{
		config:    &config.Config{},
		addresses: addrs,
		logger:    logger.NopLogger{},
	}

	for i := 0; i < 6; i++ {
		got := c.GetAddress()
		want := addrs[i%len(addrs)]
		if got != want {
			t.Errorf("call %d: got %q, want %q", i, got, want)
		}
	}
}

func TestGetAddress_SingleAddress(t *testing.T) {
	c := &Client{
		config:    &config.Config{},
		addresses: []string{"http://only:9200"},
		logger:    logger.NopLogger{},
	}

	for i := 0; i < 5; i++ {
		got := c.GetAddress()
		if got != "http://only:9200" {
			t.Errorf("call %d: expected single address, got %q", i, got)
		}
	}
}

func TestGetAddress_EmptyAddresses(t *testing.T) {
	c := &Client{
		config:    &config.Config{},
		addresses: []string{},
		logger:    logger.NopLogger{},
	}

	got := c.GetAddress()
	if got != "" {
		t.Errorf("expected empty string for no addresses, got %q", got)
	}
}
