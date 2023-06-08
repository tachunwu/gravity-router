package server

import (
	"log"

	"github.com/dgraph-io/ristretto"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	nc     *nats.Conn
	cache  *ristretto.Cache
}

func NewServer(routerURL string) *Server {
	logger, _ := zap.NewProduction()
	nc, err := nats.Connect(routerURL)
	if err != nil {
		log.Fatal(err)
	}
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // Num keys to track frequency of (10M).
		MaxCost:     1 << 30, // Maximum cost of cache (1GB).
		BufferItems: 64,      // Number of keys per Get buffer.
	})
	if err != nil {
		log.Fatal(err)
	}

	return &Server{
		logger: logger,
		nc:     nc,
		cache:  cache,
	}
}

func (s *Server) Start() {
	_, err := s.nc.Subscribe("*", func(m *nats.Msg) {
		switch m.Header.Get("op") {
		case "get":
			value, found := s.cache.Get(m.Subject)
			if found {
				s.logger.Info("get",
					zap.String("key", m.Subject),
					zap.Binary("value",
						value.([]byte)),
				)
				replyMsg := &nats.Msg{
					Subject: m.Subject,
					Data:    value.([]byte),
					Header:  make(nats.Header),
				}
				m.RespondMsg(replyMsg)
			} else {
				s.logger.Info("get: not found", zap.String("key", m.Subject))
				m.Respond([]byte("not found"))
			}
		case "set":
			s.cache.Set(m.Subject, m.Data, 1)
			s.cache.Wait()
			s.logger.Info("set",
				zap.String("key", m.Subject),
				zap.Binary("value", m.Data),
			)
			replyMsg := &nats.Msg{
				Subject: m.Subject,
				Data:    []byte("OK"),
				Header:  make(nats.Header),
			}
			m.RespondMsg(replyMsg)
		case "del":
			s.cache.Del(m.Subject)
			s.logger.Info("del",
				zap.String("key", m.Subject),
			)
			replyMsg := &nats.Msg{
				Subject: m.Subject,
				Data:    []byte("OK"),
				Header:  make(nats.Header),
			}
			m.RespondMsg(replyMsg)
		default:
			s.logger.Info("Unknown operation received",
				zap.String("op", m.Header.Get("op")),
			)
			m.Respond([]byte("unknown operation"))
		}
	})
	if err != nil {
		log.Fatal(err)
	}
}
