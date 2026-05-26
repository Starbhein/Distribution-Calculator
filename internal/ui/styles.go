package ui

import "charm.land/lipgloss/v2"

// ============================================================
// Paleta de colores — estilo cyberpunk lavanda
// ============================================================

// 🖤 Fondos
var (
	bgPrimary   = lipgloss.Color("#0D0B1E")
	bgSecondary = lipgloss.Color("#12102A")
	bgElevated  = lipgloss.Color("#1A1730")
)

// ✏️ Fuentes
var (
	textPrimary   = lipgloss.Color("#E8D5FF")
	textSecondary = lipgloss.Color("#A78BCA")
	textMuted     = lipgloss.Color("#5C4E7A")
	textAccent    = lipgloss.Color("#FF79C6")
)

// 🔤 Títulos
var (
	colTitleH1 = lipgloss.Color("#C678DD")
	colTitleH2 = lipgloss.Color("#BD93F9")
	colTitleH3 = lipgloss.Color("#8BE9FD")
)

// 📦 Bordes
var (
	borderDefault = lipgloss.Color("#3D2E6E")
	borderFocus   = lipgloss.Color("#BD93F9")
	borderActive  = lipgloss.Color("#FF79C6")
	borderDim     = lipgloss.Color("#1E1A36")
)

// 💬 Mensajes de estado
var (
	colInfo    = lipgloss.Color("#8BE9FD")
	colSuccess = lipgloss.Color("#50FA7B")
	colWarn    = lipgloss.Color("#FFB86C")
	colError   = lipgloss.Color("#FF5555")
)

// 🖱️ Otros
var (
	colTheoretical = lipgloss.Color("#BD93F9")
	colEmpirical   = lipgloss.Color("#50FA7B")
	colMark        = lipgloss.Color("#FF5555")
	colAxis        = lipgloss.Color("#5C4E7A")
	colLabel       = lipgloss.Color("#8BE9FD")
)

// ============================================================
// Estilos pre-construidos (usados en múltiples archivos)
// ============================================================

var (
	// Formulario
	buttonActiveStyle   = lipgloss.NewStyle().Foreground(colSuccess)
	buttonDisabledStyle = lipgloss.NewStyle().Foreground(colTitleH3)
	errorLabelStyle     = lipgloss.NewStyle().Foreground(colError)
	focusedStyle        = lipgloss.NewStyle().Foreground(textAccent)
	blurredStyle        = lipgloss.NewStyle().Foreground(colTitleH2)
	mutedStyle          = lipgloss.NewStyle().Foreground(textMuted)

	// Títulos y mensajes
	titleh1Style   = lipgloss.NewStyle().Foreground(colTitleH1).Bold(true)
	titleh2Style   = lipgloss.NewStyle().Foreground(colTitleH2).Bold(true)
	warningStyle   = lipgloss.NewStyle().Foreground(colWarn)
	infoStyle      = lipgloss.NewStyle().Foreground(colInfo)
	successStyle   = lipgloss.NewStyle().Foreground(colSuccess)

	// Histograma
	theoreticalStyle = lipgloss.NewStyle().Foreground(colTheoretical)
	empiricalStyle   = lipgloss.NewStyle().Foreground(colEmpirical)
	markBarStyle     = lipgloss.NewStyle().Foreground(colMark)
	axisStyle        = lipgloss.NewStyle().Foreground(colAxis)
	labelStyle       = lipgloss.NewStyle().Foreground(colLabel)

	// Spinner
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
)

// Funciones de estilo (para evitar redeclaraciones)

func secondaryTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(textSecondary)
}

func errorTextStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(colError)
}
