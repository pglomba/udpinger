package client

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

type Client struct {
	remoteName    string
	remoteAddress *net.UDPAddr
}

func NewClient(address string) (*Client, error) {
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return nil, err
	}

	return &Client{
		remoteName:    address,
		remoteAddress: udpAddr,
	}, nil
}

type Message struct {
	Id        uuid.UUID `json:"id"`
	Timestamp time.Time `json:"timestamp"`
}

func NewMessage() *Message {
	return &Message{
		Id:        uuid.New(),
		Timestamp: time.Now().UTC(),
	}
}

type RTTCheckResult struct {
	mu         sync.Mutex
	Target     string
	Results    []time.Duration
	Min        int64
	Max        int64
	Avg        int64
	PacketLoss float64
	Sent       atomic.Uint64
	Received   atomic.Uint64
}

type ConvertedRTTCheckResult struct {
	Target     string
	Results    []int64
	Min        int64
	Max        int64
	Avg        int64
	PacketLoss float64
	Sent       uint64
	Received   uint64
}

func (c *Client) RTTCheck(timeout int) (time.Duration, error) {
	conn, err := net.DialUDP("udp", nil, c.remoteAddress)
	if err != nil {
		return time.Duration(0), errors.New("error connecting to server: " + err.Error())
	}

	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(time.Duration(timeout) * time.Second))

	newMessage := NewMessage()
	encodedMessage, _ := json.Marshal(newMessage)
	_, err = conn.Write(encodedMessage)
	if err != nil {
		return time.Duration(0), errors.New("error sending message: " + err.Error())
	}

	rcvBuffer := make([]byte, len(encodedMessage))
	_, err = conn.Read(rcvBuffer)
	if err != nil {
		return time.Duration(0), errors.New("error reading message: " + err.Error())
	}

	var rcvMessage Message
	err = json.Unmarshal(rcvBuffer, &rcvMessage)
	if err != nil {
		return time.Duration(0), errors.New("error parsing message: " + err.Error())
	}

	if newMessage.Id == rcvMessage.Id {
		return time.Since(rcvMessage.Timestamp), nil
	} else {
		return time.Duration(0), errors.New("response id mismatch")
	}
}

func (c *Client) RunRTTCheck(timeout int, wg *sync.WaitGroup, result *RTTCheckResult) error {
	defer wg.Done()

	result.Sent.Add(1)

	rttCheckResult, err := c.RTTCheck(timeout)
	if err != nil {
		return err
	}

	result.Received.Add(1)
	result.mu.Lock()
	result.Results = append(result.Results, rttCheckResult)
	result.mu.Unlock()

	return nil
}

func (c *Client) StartMonitor(interval int, count int, timeout int, resultsCh chan ConvertedRTTCheckResult, timeUnit string) {
	for range time.Tick(time.Duration(interval) * time.Second) {
		var (
			rttCheckResult RTTCheckResult
			wg             sync.WaitGroup
		)

		rttCheckResult.Target = c.remoteName
		for i := 0; i < count; i++ {
			wg.Add(1)
			go func() {
				err := c.RunRTTCheck(timeout, &wg, &rttCheckResult)
				if err != nil {
					slog.Debug(err.Error())
				}
			}()
		}
		wg.Wait()

		var convertedResult ConvertedRTTCheckResult
		convertedResult.Target = rttCheckResult.Target
		convertedResult.Sent = rttCheckResult.Sent.Load()
		convertedResult.Received = rttCheckResult.Received.Load()
		for _, result := range rttCheckResult.Results {
			convertedResult.Results = append(convertedResult.Results, convertTimeUnit(result, timeUnit))
		}
		if len(convertedResult.Results) > 0 {
			convertedResult.Min = slices.Min(convertedResult.Results)
			convertedResult.Max = slices.Max(convertedResult.Results)
			convertedResult.Avg = calculateAverage(convertedResult.Results)

		}
		convertedResult.PacketLoss = calculatePacketLoss(convertedResult.Sent, convertedResult.Received)

		resultsCh <- convertedResult
	}
}

func calculateAverage(results []int64) int64 {
	var total int64
	for _, num := range results {
		total += num
	}
	return total / int64(len(results))
}

func calculatePacketLoss(sent uint64, received uint64) float64 {
	return ((float64(sent)) - float64(received)) / float64(sent) * 100
}

func convertTimeUnit(timeData time.Duration, unit string) int64 {
	switch unit {
	case "ns":
		return timeData.Nanoseconds()
	case "us":
		return timeData.Microseconds()
	case "ms":
		return timeData.Milliseconds()
	default:
		return timeData.Milliseconds()
	}
}
