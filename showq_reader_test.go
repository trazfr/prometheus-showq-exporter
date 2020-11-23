package main

import (
	"bufio"
	"bytes"
	"testing"
	"time"
)

func TestRead(t *testing.T) {
	s := ShowqReader{
		reader: bufio.NewReader(bytes.NewBufferString("item\x00")),
	}
	item, err := s.read()
	if item != "item" || err != nil {
		t.Error("item ==", item, "error ==", err)
	}
}

func TestReadKv(t *testing.T) {
	s := ShowqReader{
		reader: bufio.NewReader(bytes.NewBufferString("key\x00value\x00")),
	}
	key, value, err := s.readkv()
	if key != "key" || value != "value" || err != nil {
		t.Error("key ==", key, "|| value ==", value, "error ==", err)
	}
}

func TestAll(t *testing.T) {
	payload := "queue_name\x00deferred\x00queue_id\x00362126BB12\x00time\x001603352331\x00size\x0056772\x00sender\x00from@example.com\x00recipient\x00to@example.org\x00reason\x00connect to example.org[1.2.3.4]:25: Connection timed out\x00\x00" +
		"queue_name\x00deferred\x00queue_id\x00CB2D66ABBE\x00time\x001603350152\x00size\x0051486\x00sender\x00from2@example.com\x00recipient\x00to2@example.org\x00reason\x00connect to example.org[1.2.3.4]:25: Connection timed out\x00\x00\x00"
	s := NewShowqReader(bytes.NewBufferString(payload))

	{
		item, ok := s.ReadItem()
		if !ok || item.queueName != "deferred" ||
			item.queueID != "362126BB12" ||
			item.time != time.Unix(1603352331, 0) ||
			item.size != 56772 ||
			item.sender != "from@example.com" ||
			item.recipient != "to@example.org" ||
			item.reason != "connect to example.org[1.2.3.4]:25: Connection timed out" {
			t.Error("ok =", ok, "wrong item:", item)
		}
	}
	{
		item, ok := s.ReadItem()
		if !ok || item.queueName != "deferred" ||
			item.queueID != "CB2D66ABBE" ||
			item.time != time.Unix(1603350152, 0) ||
			item.size != 51486 ||
			item.sender != "from2@example.com" ||
			item.recipient != "to2@example.org" ||
			item.reason != "connect to example.org[1.2.3.4]:25: Connection timed out" {
			t.Error("ok =", ok, "wrong item:", item)
		}
	}

	if _, ok := s.ReadItem(); ok {
		t.Error("ok =", ok)
	}
}
