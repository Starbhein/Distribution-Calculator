# DistCalc — Simulador de Distribuciones de Probabilidad

> Calculadora interactiva de distribuciones de probabilidad con simulación Monte Carlo, visualización en terminal y exportación de resultados.

## 🚀 Características

- **9 distribuciones** implementadas (5 discretas + 4 continuas)
- **Simulación Monte Carlo** con 4 workers concurrentes
- **Comparación teórica vs empírica** — media, varianza, desv. estándar
- **Probabilidades** — P(X=x)/f(x), P(X≤x), P(X>x)
- **Histogramas ASCII** con marcador del valor x en la TUI
- **Exportación** de gráficas a PNG/SVG (con curva PDF teórica) y datos a CSV
- **TUI** con Bubble Tea v2

## 📦 Instalación

### Requisitos

- Go 1.26 o superior
- Terminal con soporte de colores ANSI (recomendado)

### Pasos

```bash
# Clonar el repositorio
git clone https://github.com/Starbhein/Distribution-Calculator.git
cd Distribution-Calculator

# Descargar dependencias
go mod download

# Compilar
go build ./cmd/tui/main.go

# O ejecutar directamente
go run ./cmd/tui/main.go
```

## 🎮 Uso

### Menú principal

Al iniciar, verás un menú con las 9 distribuciones disponibles. Usa **↑↓** para navegar y **Enter** para seleccionar.

### Formulario de parámetros

Según la distribución elegida, se muestran inputs dinámicos:

| Distribución      | Parámetros                                  |
| ----------------- | ------------------------------------------- |
| Bernoulli         | p (probabilidad), x                         |
| Binomial          | p, n (ensayos), x                           |
| Poisson           | λ (lambda), x                               |
| Geométrica        | p, x                                        |
| Hipergeométrica   | N (población), M (éxitos), n (muestra), x   |
| Normal            | μ (media), σ (desv. estándar), x            |
| Exponencial (λ)   | λ, x                                        |
| Exponencial (β)   | β, x                                        |
| Uniforme continua | a (límite inferior), b (límite superior), x |

El último parámetro siempre es **x** (valor a evaluar) y se agrega automáticamente el **tamaño de muestra** (default: 1000).

### Navegación

| Tecla           | Acción                     |
| --------------- | -------------------------- |
| Tab / Shift+Tab | Siguiente / anterior input |
| Enter           | Simular                    |
| ESC             | Volver al menú             |
| CTRL+C          | Salir                      |

### Resultados

Tras la simulación se muestra:

- **Tabla comparativa** — teórico vs empírico (media, varianza, desv. estándar)
- **Probabilidades** — P(X=x)/f(x), P(X≤x), P(X>x)
- **Histograma ASCII** — con barras proporcionales y marcador del valor x

### Exportación

| Tecla | Formato | Descripción                                |
| ----- | ------- | ------------------------------------------ |
| `e`   | PNG     | Gráfica con histograma + curva PDF teórica |
| `s`   | SVG     | Misma gráfica en formato vectorial         |
| `c`   | CSV     | Datos crudos simulados (índice, valor)     |

Los archivos se guardan en el directorio actual con nombres auto-generados: `distcalc-<distribución>-<timestamp>.<ext>`

**Elementos de la gráfica exportada:**

- Título con distribución, n, μ̂, σ̂ y parámetros
- Ejes: Valor (X) / Densidad (Y)
- Leyenda: "Histograma empírico", "PDF teórica", "x = valor"
- Histograma normalizado (área = 1)
- Curva PDF teórica superpuesta
- Línea vertical punteada en el valor x

## 📊 Distribuciones soportadas

### Discretas

| Distribución    | PMF                     | Media | Varianza                 |
| --------------- | ----------------------- | ----- | ------------------------ |
| Bernoulli       | pᵏ(1-p)¹⁻ᵏ              | p     | p(1-p)                   |
| Binomial        | C(n,k)pᵏ(1-p)ⁿ⁻ᵏ        | np    | np(1-p)                  |
| Poisson         | e⁻ˡλᵏ/k!                | λ     | λ                        |
| Geométrica      | (1-p)ᵏ⁻¹p               | 1/p   | (1-p)/p²                 |
| Hipergeométrica | C(M,k)C(N-M,n-k)/C(N,n) | nM/N  | n(M/N)(1-M/N)(N-n)/(N-1) |

### Continuas

| Distribución      | PDF                  | Media   | Varianza  |
| ----------------- | -------------------- | ------- | --------- |
| Normal            | (1/σ√2π)e⁻⁽ˣ⁻μ⁾²/²σ² | μ       | σ²        |
| Exponencial (λ)   | λe⁻ˡˣ                | 1/λ     | 1/λ²      |
| Exponencial (β)   | (1/β)e⁻ˣ/β           | β       | β²        |
| Uniforme continua | 1/(b-a)              | (a+b)/2 | (b-a)²/12 |

## 🏗️ Arquitectura

```
cmd/tui/          # Punto de entrada
internal/
  core/
    distributions/    # Estructuras y operaciones teóricas
    distributions/sim/# Motor de simulación Monte Carlo
    stats/            # Estadísticos empíricos (Welford)
  ui/               # TUI con Bubble Tea
  export/           # Exportación PNG/SVG/CSV
```

### Optimizaciones de simulación

- **Tablas CDF compartidas** entre workers (Binomial, Poisson, Hipergeométrica)
- **Aproximación normal** cuando varianza > 9
- **Método híbrido** para Geométrica (transformada inversa O(1) para p≥0.01, iterativo para p<0.01)
- **Precisión de punto flotante**: log-gamma para factoriales, `math.Expm1` para cancelación catastrófica

## 🛠️ Stack tecnológico

- **Go 1.26** — lenguaje principal
- **Bubble Tea v2** — framework TUI
- **Lipgloss v2** — estilos y colores
- **Bubbles v2** — componentes (spinner, list, textinput)
- **gonum/plot** — gráficas PNG/SVG
- **math/rand/v2** — PRNG con PCG

## 📄 Licencia

MIT — Ver [LICENSE](LICENSE) para más detalles.

---
