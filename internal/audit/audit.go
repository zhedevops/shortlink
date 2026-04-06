package audit

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

const (
	ActionShorten = "shorten"
	ActionFollow  = "follow"
)

type AuditService struct {
	ch    chan AuditEvent
	sinks []AuditSink
}

type AuditEvent struct {
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`
	URL       string    `json:"url"`
	Timestamp time.Time `json:"ts"`
}

type AuditSink interface {
	Consume(event AuditEvent)
}

type FileSink struct {
	mu sync.Mutex // для безопасной записи из нескольких горутин

	Path string
}

func NewFileSink(path string) *FileSink {
	return &FileSink{mu: sync.Mutex{}, Path: path}
}

type RemoteSink struct {
	client *resty.Client

	URL string
}

func NewRemoteSink(url string) *RemoteSink {
	return &RemoteSink{client: resty.New(), URL: url}
}

func NewAuditService(sinks []AuditSink) *AuditService {
	svc := &AuditService{
		ch:    make(chan AuditEvent, 20),
		sinks: sinks,
	}
	go svc.run()
	return svc
}

func (a *AuditService) run() {
	for e := range a.ch {
		for _, s := range a.sinks {
			go s.Consume(e) // каждый sink в отдельной горутине, чтобы один не блокировал остальных
		}
	}
}

func (a *AuditService) Send(e AuditEvent) {
	select {
	case a.ch <- e:
	default:
		log.Printf("audit: channel full, event dropped")
	}
}

func (fs *FileSink) Consume(e AuditEvent) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	file, err := os.OpenFile(fs.Path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Error().Err(err).Msg("audit: open file failed")
		return
	}
	defer func() {
		_ = file.Close()
	}()

	data, err := json.Marshal(e)
	if err != nil {
		log.Error().Err(err).Msg("audit: marshal failed")
		return
	}

	if _, err = file.Write(append(data, '\n')); err != nil {
		log.Error().Err(err).Msg("audit: write failed")
		return
	}
}

func (rs *RemoteSink) Consume(e AuditEvent) {
	resp, err := rs.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(e).
		Post(rs.URL)
	if err != nil {
		log.Error().Err(err).Msg("audit: request failed")
		return
	}
	if resp.IsError() {
		log.Error().
			Int("status", resp.StatusCode()).
			Str("body", resp.String()).
			Msg("audit: server returned error")
	}
}
