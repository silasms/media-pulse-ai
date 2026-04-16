package ai

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/silasms/media-pulse-ai/internal/domain"
)

type MockProvider struct {
	mu          sync.Mutex
	failNext    bool
	delay       time.Duration
	callCounter int
}

func NewMockProvider() *MockProvider {
	return &MockProvider{
		delay: 10 * time.Millisecond,
	}
}

func (m *MockProvider) SetFailNext(fail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failNext = fail
}

func (m *MockProvider) Name() string {
	return "mock-ai-engine-v1"
}

func (m *MockProvider) Enrich(ctx context.Context, input EnrichmentInput) (*domain.EnrichedMetadata, error) {
	m.mu.Lock()
	if m.failNext {
		m.failNext = false
		m.mu.Unlock()
		return nil, fmt.Errorf("mock AI provider temporary service failure")
	}
	m.callCounter++
	delay := m.delay
	m.mu.Unlock()

	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	lowerTitle := strings.ToLower(input.Title)
	lowerDesc := strings.ToLower(input.Description)

	themes := []string{"Audiovisual", "Entretenimento"}
	rating := domain.RatingGeneral
	var warnings []string
	sentiment := "Neutro"

	if strings.Contains(lowerTitle, "tech") || strings.Contains(lowerTitle, "sistemas") || strings.Contains(lowerDesc, "dados") {
		themes = append(themes, "Tecnologia", "Sistemas Distribuídos")
		sentiment = "Informativo"
	}
	if strings.Contains(lowerTitle, "futebol") || strings.Contains(lowerTitle, "esporte") || strings.Contains(lowerDesc, "gol") {
		themes = append(themes, "Esportes", "Futebol")
		sentiment = "Empolgante"
	}
	if strings.Contains(lowerTitle, "ação") || strings.Contains(lowerDesc, "tiro") || strings.Contains(lowerDesc, "violência") {
		rating = domain.Rating14
		warnings = append(warnings, "Violência moderada")
		sentiment = "Tenso"
	}

	chapters := make([]domain.Chapter, 0)
	step := input.DurationSeconds / 3
	if step < 60 {
		step = input.DurationSeconds
	}

	for i := 0; i < input.DurationSeconds; i += step {
		end := i + step
		if end > input.DurationSeconds {
			end = input.DurationSeconds
		}
		chNum := len(chapters) + 1
		chapters = append(chapters, domain.Chapter{
			StartSecond: i,
			EndSecond:   end,
			Title:       fmt.Sprintf("Parte %d: Destaques de %s", chNum, input.Title),
			Summary:     fmt.Sprintf("Segmento cobrindo o bloco principal de %s.", input.Title),
		})
	}

	summary := fmt.Sprintf("Conteúdo completo de '%s'. Destaque para narrativa dinâmica e abordagem temática de alta relevância.", input.Title)
	if input.Description != "" {
		summary = fmt.Sprintf("%s - %s", input.Title, input.Description)
	}

	tags := append(themes, strings.Split(lowerTitle, " ")...)

	return &domain.EnrichedMetadata{
		Summary:         summary,
		KeyThemes:       themes,
		Chapters:        chapters,
		ParentalRating:  rating,
		ContentWarnings: warnings,
		Tags:            tags,
		Sentiment:       sentiment,
		ModelUsed:       m.Name(),
		ProcessedAt:     time.Now().UTC(),
	}, nil
}

func (m *MockProvider) GenerateEmbeddings(ctx context.Context, text string) ([]float32, error) {
	dims := 384
	vec := make([]float32, dims)
	h := float64(len(text))
	for i := 0; i < dims; i++ {
		val := math.Sin(h*float64(i+1)) * 0.1
		vec[i] = float32(val)
	}
	return vec, nil
}
