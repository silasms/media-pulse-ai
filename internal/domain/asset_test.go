package domain_test

import (
	"testing"

	"github.com/silasms/media-pulse-ai/internal/domain"
)

func TestMediaAsset_Validate(t *testing.T) {
	tests := []struct {
		name     string
		asset    domain.MediaAsset
		wantErr  error
		wantLang string
	}{
		{
			name: "valid asset with default language",
			asset: domain.MediaAsset{
				Title:           "Tech Talks - Edição Especial",
				MediaURL:        "https://cdn.example.com/vod/tech-talks.mp4",
				DurationSeconds: 2700,
			},
			wantErr:  nil,
			wantLang: "pt-BR",
		},
		{
			name: "valid asset with custom language",
			asset: domain.MediaAsset{
				Title:           "Special Documentary",
				MediaURL:        "https://cdn.example.com/vod/doc.mp4",
				DurationSeconds: 1800,
				Language:        "en-US",
			},
			wantErr:  nil,
			wantLang: "en-US",
		},
		{
			name: "missing title",
			asset: domain.MediaAsset{
				Title:           "",
				MediaURL:        "https://cdn.example.com/vod/item.mp4",
				DurationSeconds: 600,
			},
			wantErr: domain.ErrInvalidTitle,
		},
		{
			name: "missing media URL",
			asset: domain.MediaAsset{
				Title:           "Resumo da Rodada",
				MediaURL:        "",
				DurationSeconds: 600,
			},
			wantErr: domain.ErrInvalidMediaURL,
		},
		{
			name: "invalid duration zero",
			asset: domain.MediaAsset{
				Title:           "Podcast Resumo",
				MediaURL:        "https://cdn.example.com/podcasts/ep1.mp3",
				DurationSeconds: 0,
			},
			wantErr: domain.ErrInvalidDuration,
		},
		{
			name: "invalid negative duration",
			asset: domain.MediaAsset{
				Title:           "Podcast Resumo",
				MediaURL:        "https://cdn.example.com/podcasts/ep1.mp3",
				DurationSeconds: -100,
			},
			wantErr: domain.ErrInvalidDuration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.asset.Validate()
			if err != tt.wantErr {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if err == nil && tt.asset.Language != tt.wantLang {
				t.Fatalf("expected language %s, got %s", tt.wantLang, tt.asset.Language)
			}
		})
	}
}
