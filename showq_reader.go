package main

import (
	"bufio"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type ShowqReader struct {
	reader *bufio.Reader
	ch     chan *ShowqItem
}

type ShowqItem struct {
	queueName string
	queueID   string
	time      time.Time
	size      uint
	sender    string
	recipient string
	reason    string
	inError   bool
}

type kv struct {
	k string
	v string
}

func NewShowqReader(rd io.Reader) ShowqReader {
	result := ShowqReader{
		reader: bufio.NewReader(rd),
		ch:     make(chan *ShowqItem),
	}
	go result.run()
	return result
}

func (s *ShowqReader) ReadItem() (*ShowqItem, bool) {
	res, ok := <-s.ch
	return res, ok
}

func (s *ShowqReader) run() {
	for {
		kvRead := false
		item := &ShowqItem{}

		for {
			k, v, err := s.readkv()
			if err != nil {
				log.Println("ERROR:", err)
				item.inError = true
				s.ch <- item
				close(s.ch)
				return
			}
			if k == "" && v == "" {
				break
			}

			switch k {
			case "queue_name":
				item.queueName = v
			case "queue_id":
				item.queueID = v
			case "time":
				timestamp, _ := strconv.ParseInt(v, 10, 64)
				item.time = time.Unix(timestamp, 0)
			case "size":
				size, _ := strconv.ParseUint(v, 10, 64)
				item.size = uint(size)
			case "sender":
				item.sender = v
			case "recipient":
				item.recipient = v
			case "reason":
				item.reason = v
			default:
			}

			kvRead = true
		}

		if kvRead {
			s.ch <- item
		} else {
			break
		}
	}
	close(s.ch)
}

func (s *ShowqReader) read() (string, error) {
	str, err := s.reader.ReadString('\x00')
	if err == nil {
		str = strings.TrimRight(str, "\x00")
	}
	return str, err
}

func (s *ShowqReader) readkv() (string, string, error) {
	k, err := s.read()
	if err != nil || k == "" {
		return "", "", err
	}
	v, err := s.read()
	return k, v, err
}
