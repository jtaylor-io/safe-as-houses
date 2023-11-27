package worker

import (
	"context"

	"github.com/hibiken/asynq"
	db "github.com/jtaylor-io/safe-as-houses/db/sqlc"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	QueueCritical = "critical"
	QueueDefault  = "default"
)

type TaskProcessor interface {
	Start() error
	ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error
}

type RedisTaskProcessor struct {
	server *asynq.Server
	store  db.Store
}

func NewRedisTaskProcessor(redisOpt asynq.RedisClientOpt, store db.Store) TaskProcessor {
	logger := NewLogger()
	redis.SetLogger(logger)
	server := asynq.NewServer(
		redisOpt,
		asynq.Config{
			Queues: map[string]int{
				QueueCritical: 10,
				QueueDefault:  5,
			},
			ErrorHandler: asynq.ErrorHandlerFunc(
				func(ctx context.Context, task *asynq.Task, err error) {
					log.Error().
						Err(err).
						Str("type", task.Type()).
						Bytes("payload", task.Payload()).
						Msg("prosess task failed")
				},
			),
			Logger: logger,
		},
	)
	return &RedisTaskProcessor{
		server,
		store,
	}
}
