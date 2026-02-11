package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const canonicalProgressBarWidth = 24

func formatProgressLabel(done, total int) string {
	if total > 0 {
		pct := int(float64(min(done, total)) / float64(total) * 100)
		return fmt.Sprintf("%d/%d  %d%%", done, total, pct)
	}
	return fmt.Sprintf("%d seen", done)
}

func renderCanonicalProgressBar(done, total, frame int) string {
	return renderCanonicalProgressBarWidth(done, total, canonicalProgressBarWidth, frame)
}

func renderCanonicalProgressBarWidth(done, total, width, frame int) string {
	if width <= 0 {
		width = 1
	}

	fillInline := lipgloss.NewStyle().Foreground(secondaryColor).Bold(true)
	emptyInline := lipgloss.NewStyle().Foreground(borderColor)
	pulseInline := lipgloss.NewStyle().Foreground(primaryColor).Bold(true)

	if total <= 0 {
		pulse := frame % width
		bar := strings.Repeat("░", width)
		runes := []rune(bar)
		runes[pulse] = '▣'
		return pulseInline.Render(string(runes))
	}

	if done < 0 {
		done = 0
	}
	if done > total {
		done = total
	}

	filled := int(float64(done) / float64(total) * float64(width))
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}

	filledPart := strings.Repeat("█", filled)
	emptyPart := strings.Repeat("░", width-filled)
	return fillInline.Render(filledPart) + emptyInline.Render(emptyPart)
}
