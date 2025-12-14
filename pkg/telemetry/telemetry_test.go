package telemetry

import (
	"testing"
	"time"
)

func TestString(t *testing.T) {
	attr := String("key", "value")
	if attr.Key != "key" {
		t.Errorf("expected key 'key', got '%s'", attr.Key)
	}
	if attr.Value != "value" {
		t.Errorf("expected value 'value', got '%v'", attr.Value)
	}
}

func TestInt(t *testing.T) {
	attr := Int("count", 42)
	if attr.Key != "count" {
		t.Errorf("expected key 'count', got '%s'", attr.Key)
	}
	if attr.Value != 42 {
		t.Errorf("expected value 42, got %v", attr.Value)
	}
}

func TestInt64(t *testing.T) {
	attr := Int64("bignum", int64(9223372036854775807))
	if attr.Key != "bignum" {
		t.Errorf("expected key 'bignum', got '%s'", attr.Key)
	}
	if attr.Value != int64(9223372036854775807) {
		t.Errorf("expected value 9223372036854775807, got %v", attr.Value)
	}
}

func TestBool(t *testing.T) {
	attr := Bool("enabled", true)
	if attr.Key != "enabled" {
		t.Errorf("expected key 'enabled', got '%s'", attr.Key)
	}
	if attr.Value != true {
		t.Errorf("expected value true, got %v", attr.Value)
	}
}

func TestFloat64(t *testing.T) {
	attr := Float64("rate", 3.14159)
	if attr.Key != "rate" {
		t.Errorf("expected key 'rate', got '%s'", attr.Key)
	}
	if attr.Value != 3.14159 {
		t.Errorf("expected value 3.14159, got %v", attr.Value)
	}
}

func TestDuration(t *testing.T) {
	d := 100 * time.Millisecond
	attr := Duration("latency", d)
	if attr.Key != "latency" {
		t.Errorf("expected key 'latency', got '%s'", attr.Key)
	}
	if attr.Value != int64(100) {
		t.Errorf("expected value 100ms, got %v", attr.Value)
	}
}

func TestWithAttributes(t *testing.T) {
	config := &SpanConfig{}
	opt := WithAttributes(String("key1", "val1"), Int("key2", 42))
	opt(config)

	if len(config.Attributes) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(config.Attributes))
	}
	if config.Attributes[0].Key != "key1" || config.Attributes[0].Value != "val1" {
		t.Errorf("unexpected first attribute: %+v", config.Attributes[0])
	}
	if config.Attributes[1].Key != "key2" || config.Attributes[1].Value != 42 {
		t.Errorf("unexpected second attribute: %+v", config.Attributes[1])
	}
}

func TestWithAttributesMultipleCalls(t *testing.T) {
	config := &SpanConfig{}
	opt1 := WithAttributes(String("key1", "val1"))
	opt2 := WithAttributes(String("key2", "val2"))

	opt1(config)
	opt2(config)

	if len(config.Attributes) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(config.Attributes))
	}
}
