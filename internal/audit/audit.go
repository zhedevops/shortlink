package audit

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

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
	Path string
	Mu   sync.Mutex // для безопасной записи из нескольких горутин
}

type RemoteSink struct {
	URL    string
	Client *http.Client
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
	fs.Mu.Lock()
	defer fs.Mu.Unlock()

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

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		log.Error().Err(err).Msg("audit: write failed")
		return
	}
}

func (rs *RemoteSink) Consume(e AuditEvent) {
	data, err := json.Marshal(e)
	if err != nil {
		log.Error().Err(err).Msg("audit: marshal failed")
		return
	}
	req, err := http.NewRequest(http.MethodPost, rs.URL, bytes.NewBuffer(data))
	if err != nil {
		log.Error().Err(err).Msg("audit: request create failed")
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := rs.Client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("audit: request failed")
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
}
