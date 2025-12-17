package executor

import (
	"bytes"
	"context"
	"fmt"
	"job_scheduler_go_rabbitmq/internal/core/domain"
	"net/http"
)

type HTTPExecutor struct {
	client *http.Client
}

func NewHTTPExecutor() *HTTPExecutor {
	return &HTTPExecutor{
		client: &http.Client{},
	}
}

func (e *HTTPExecutor) Execute(ctx context.Context, job *domain.Job) domain.ExecutionResult {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		job.CallbackURL,
		bytes.NewReader(job.Payload),
	)
	if err != nil {
		return domain.ExecutionResult{
			HTTPStatus: 0,
			Error:      err,
		}
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return domain.ExecutionResult{
			HTTPStatus: 0,
			Error:      err,
		}
	}
	defer resp.Body.Close()

	// Siempre capturamos el status
	if resp.StatusCode >= 400 {
		return domain.ExecutionResult{
			HTTPStatus: resp.StatusCode,
			Error:      fmt.Errorf("callback failed with status %d", resp.StatusCode),
		}
	}

	return domain.ExecutionResult{
		HTTPStatus: resp.StatusCode,
		Error:      nil,
	}
}
