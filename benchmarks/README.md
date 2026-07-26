# Benchmarks: evidencia de precisión y escalado algorítmico

`compare.py` produce dos evidencias independientes, 100% stdlib de Python:
(1) una **tabla de precisión** que contrasta referencias racionales exactas
(`math.comb` + `fractions.Fraction`) contra las recurrencias float64 que usa
el núcleo `distmath` de Go, portadas línea por línea; y (2) un **barrido de
escalado** que compara CLASES DE ALGORITMOS (nunca lenguajes) midiendo
exponentes de crecimiento.

## Camino rápido

1. Requisito: **Python 3.8+** (`math.comb`); sin dependencias externas ni
   `requirements.txt`.
2. Ejecutar: `python3 benchmarks/compare.py` (tarda segundos; el tramo
   geométrico p=0.0001 hace 2e7 iteraciones).
3. Verificar: los valores exactos impresos deben coincidir con la checklist
   de abajo. `make bench-compare` es un atajo equivalente con guarda de
   `python3`.

## Método

| Parte | Qué mide | Por qué es honesta |
|-------|----------|--------------------|
| Precisión | Drift abs/rel del float64 vs racionales exactos | La referencia es aritmética exacta, no otro float |
| Barrido hipergeométrico | Recurrencia de fila O(S) vs forma cerrada ingenua por punto O(S²) | Ambos en Python float64: misma máquina, mismo intérprete |
| Barrido geométrico | Inversa-CDF O(1) (`k = ⌈log(u)/log1p(−p)⌉`) vs bucle de ensayos O(1/p) | Pendiente log-log por mínimos cuadrados (stdlib) |

La forma cerrada ingenua es la multiplicativa directa O(k) por punto
(obligatoria por diseño): una variante con `lgamma` por punto sería lineal
con constante grande y NO mostraría la diferencia de clase. Los tres
factores C(·,·) se evalúan como un solo producto intercalado para que los
intermedios no desborden float64 en N=1600; el costo por punto sigue O(k).

## Tamaños elegidos y justificación

| Caso | Tamaño | Justificación |
|------|--------|---------------|
| Hipergeométrica (precisión) | M=3, N=12, n=4 | Caso fijado en los tests de Go (PMF(0)=14/55) |
| Hipergeométrica (precisión) | N=50, M=20, n=10 | Caso realista del simulador; C(50,10)≈1e10, trivial para enteros exactos |
| Binomial (precisión) | n=10, p=0.1 | Caso fijado en los tests de Go (PMF(2)=0.1937102445) |
| Binomial (precisión) | n=1000, p=0.9 | Caso del BenchmarkCDF de Go; C(1000,k)≈300 dígitos, aún rápido |
| Barrido hipergeométrico | N ∈ {200,400,800,1600}, M=N/2, n=N/4 → S ∈ {51,101,201,401} | Duplicar N separa pendientes ≈1 vs ≈2; S=401 mantiene el tramo cuadrático en milisegundos |
| Barrido geométrico | m=2000, p ∈ {0.1, 0.01, 0.001, 0.0001} | Plano vs ×10 por paso (pendiente ≈0 vs ≈1 contra 1/p); peor tramo m/p = 2e7 iteraciones ≈ segundos |

## Checklist de corrección reproducible

Verificación MANUAL (sin automatización — YAGNI hasta que haya CI): ejecutar
`python3 benchmarks/compare.py` y confirmar que los valores exactos impresos
coinciden con los valores fijados en los tests de Go:

- [ ] Hipergeométrica M=3, N=12, n=4: PMF(0) exacta = **14/55** ≈ 0.2545454545454545
  (pin Go: `TestHypergeometricPMFPinnedValues`, `internal/core/distmath/distmath_test.go`)
- [ ] Hipergeométrica M=3, N=12, n=4: CDF(2) exacta = **54/55** ≈ 0.9818181818181818
  (pin Go: `TestHypergeometricCDFPinnedValues`, `internal/core/distmath/cdf_test.go`)
- [ ] Binomial n=10, p=0.1: PMF(2) exacta = 387420489/2000000000 = **0.1937102445**
  (pin Go: `TestBinomial`, `internal/core/distributions/binomial_test.go`)
- [ ] Binomial n=10, p=0.1: CDF(2) exacta = 1162261467/1250000000 = **0.9298091736**
  (pin Go: `TestBinomialCDFPinnedValues`, suma PMF(0..2), `cdf_test.go`)
- [ ] Binomial n=1000, p=0.9: CDF(999) exacta = 1 − 9¹⁰⁰⁰/10¹⁰⁰⁰ ≈ 1 − 1.75e-46;
  la recurrencia float64 debe caer dentro de 1e-12 de 1
  (pin Go: `TestBinomialCDFPinnedValues`, caso benchmark, `cdf_test.go`)
- [ ] Todas las filas exactas suman exactamente 1; las filas float64 suman 1
  dentro de ~1 ulp.

Nota: el caso hipergeométrico N=50, M=20, n=10 es un caso de precisión del
diseño, NO un pin de Go — su PMF(0) exacta es C(20,0)·C(30,10)/C(50,10) =
30045015/10272278170 = 3393/1160054 ≈ 0.0029248638425452608, distinta del
pin 14/55 (que vive en M=3, N=12, n=4).

## Qué NO se afirma

- Ningún ns/op absoluto es comparable entre máquinas: los tiempos son de ESTA
  máquina y solo importan las pendientes de crecimiento.
- Ningún titular "Go > Python": el script nunca ejecuta ni parsea Go; los
  tiempos de Go salen de `make bench` y la correlación es manual.
- Ruido de timing: mediana de ≥3 repeticiones (`statistics.median`), pero en
   máquinas compartidas los tramos pequeños pueden variar; las pendientes
  log-log son la señal robusta.
- El drift relativo máximo se reporta sobre masa ≥ 1e-15: el error relativo
  en el polvo de la cola lejana es ruido, no señal.

## Siguiente paso

Leer junto a `make bench` (benchmarks de Go con etiquetas de procedencia) y
la sección «⚡ Rendimiento y precisión» del README raíz.
