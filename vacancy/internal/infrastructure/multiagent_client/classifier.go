package multiagent_client

import (
	"context"
	"fmt"

	pb "github.com/artem13815/hr/vacancy/internal/pb/multiagent_api"
	"github.com/artem13815/hr/vacancy/internal/usecase"
)

// Classifier implements usecase.RoleClassifier by calling
// multiagent.ClassifyRole over gRPC. Stateless and safe to share across
// goroutines — the underlying client multiplexes RPCs on a single HTTP/2
// connection.
type Classifier struct {
	client pb.MultiAgentServiceClient
}

// NewClassifier wraps an existing multiagent gRPC client. We accept the
// client (not a config) so the same conn is reused across constructors —
// dialing happens exactly once in bootstrap.
func NewClassifier(client pb.MultiAgentServiceClient) *Classifier {
	return &Classifier{client: client}
}

var _ usecase.RoleClassifier = (*Classifier)(nil)

// Classify maps a vacancy onto a multiagent prompt role. Any RPC failure
// (Unavailable, DeadlineExceeded, Internal, InvalidArgument, …) is wrapped
// with usecase.ErrLLMUnavailable so the usecase can take exactly one
// fallback path — the keyword detector. Vacancy CRUD never blocks on a
// flaky classifier; the trade-off is acceptable because the keyword
// fallback is always in-vocabulary.
func (c *Classifier) Classify(ctx context.Context, title, description string) (string, error) {
	resp, err := c.client.ClassifyRole(ctx, &pb.ClassifyRoleRequest{
		Title:       title,
		Description: description,
	})
	if err != nil {
		return "", fmt.Errorf("%w: %v", usecase.ErrLLMUnavailable, err)
	}
	return resp.GetRole(), nil
}
