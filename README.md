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
- **Geométrica por transformada inversa O(1)** para todo p (`k = ⌈log(u)/log1p(−p)⌉`, un draw por muestra)
- **Precisión de punto flotante**: log-gamma para factoriales, `math.Expm1` para cancelación catastrófica

## ⚡ Rendimiento y precisión

### Resultados archivados

Speedups medidos al momento del merge en la máquina de desarrollo; **baseline = commit previo a cada PR**. No son promesas absolutas entre máquinas (ver «Qué NO afirmamos»).

| Speedup | Algoritmo | Procedencia |
| ------- | --------- | ----------- |
| ≈11× | Barrido del soporte hipergeométrico (~4 920 → ~436 ns/op) | Medido al mergear PR #10; baseline = commit previo al merge (ver descripción del PR y reporte de archivo `optimize-distribution-processing`) |
| ≈284× | `FillGeometric` p=0.001, 1e6 muestras (~2.197 s → ~7.74 ms) | Medido al mergear PR #10; baseline = commit previo al merge (misma fuente) |
| ≈4.1× | CDF binomial (n=1000, ~888 → ~214.9 ns/op) | Medido al mergear PR #11; baseline = commit previo al merge (misma fuente) |

### Cómo reproducir

```bash
make bench           # benchmarks de Go, flags fijados (-benchtime=1s -count=5)
make bench-precision # valores fijados + triangulación + momentos 1e5 y 1e7 (pesado)
make bench-compare   # barrido Python stdlib: clases de algoritmos (opcional; necesita python3)
```

`make bench-precision` exporta `DISTCALC_HEAVY=1` para incluir la variante pesada de 1e7 muestras; en `make test` y `make test-short` ese test se **omite** siempre.

### Qué NO afirmamos

- **Sin ns/op absolutos entre máquinas**: los números archivados son de UNA máquina en el momento del merge; la señal robusta son las pendientes de crecimiento y los órdenes de magnitud.
- **Sin titular «Go > Python»**: `benchmarks/compare.py` nunca ejecuta ni parsea Go; compara CLASES DE ALGORITMOS dentro de Python (inversa-CDF O(1) vs bucle de ensayos O(1/p); recurrencia de fila O(rango) vs forma cerrada ingenua por punto). Los tiempos de Go salen de `make bench` y la correlación es manual.
- **Sin gates de CI**: todos estos targets son informativos; ningún benchmark bloquea merges.

### Método y caveats

- **Flags fijados**: `-benchtime=1s -count=5` en `make bench`; Go 1.26 (ver `go.mod`).
- **Precisión**: valores fijados con tolerancia 1e-12 en `internal/core/distmath` (PMF/CDF hipergeométrica, binomial y poisson), verificación cruzada contra referencias de racionales exactos (`math.comb` + `fractions.Fraction`) vía `make bench-compare`, y tests de momentos geométricos contra los valores teóricos (1e5 muestras en todo run; 1e7 solo con `DISTCALC_HEAVY=1`).
- **Forma del barrido reproducible**: los exponentes de crecimiento (lineal vs cuadrático en hipergeométrica; plano vs ∝1/p en geométrica) se reproducen en cualquier máquina con `make bench-compare`. Método, tamaños y checklist de corrección: `benchmarks/README.md`.
- **Ruido**: en máquinas compartidas las piernas pequeñas varían; `make bench` fija `-count=5` y `compare.py` usa mediana de ≥3 repeticiones, pero los ns/op individuales pueden moverse.

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
