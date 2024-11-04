package processor

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/autobrr/omegabrr/internal/domain"
)

func TestService_shouldProcessItem(t *testing.T) {
	s := &Service{}
	cfg := &domain.ArrConfig{
		IncludeUnmonitored: true,
	}
	assert.True(t, s.shouldProcessItem(false, cfg), "unmonitored items should be processed when includeUnmonitored is true")
}
