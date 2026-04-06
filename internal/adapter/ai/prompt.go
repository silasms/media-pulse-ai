package ai

import (
	"fmt"
	"strings"
)

const SystemPrompt = `Você é um especialista em análise e catalogação de conteúdo audiovisual.
Sua tarefa é analisar o título, descrição, duração e transcrição fornecidos e retornar uma análise estruturada contendo:
1. Resumo executivo (synopsis).
2. Temas-chave e palavras-chave para catálogo.
3. Capítulos/cenas com minutagem estimada (start_second, end_second, title, summary).
4. Classificação Indicativa recomendada para o público (L, 10, 12, 14, 16, 18).
5. Alertas de conteúdo (violência, conteúdo sexual, drogas lícitas/ilícitas).
6. Sentimento e tom narrativo dominante.`

func FormatEnrichmentPrompt(input EnrichmentInput) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Título: %s\n", input.Title))
	sb.WriteString(fmt.Sprintf("Descrição: %s\n", input.Description))
	sb.WriteString(fmt.Sprintf("Duração: %d segundos\n", input.DurationSeconds))
	sb.WriteString(fmt.Sprintf("Idioma: %s\n", input.Language))

	if input.Transcript != "" {
		sb.WriteString(fmt.Sprintf("Transcrição de Áudio:\n%s\n", input.Transcript))
	} else {
		sb.WriteString("Transcrição: Não fornecida (analise com base nos metadados de título e descrição).\n")
	}

	return sb.String()
}
