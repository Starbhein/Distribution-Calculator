# Simulador de Distribuciones de Probabilidad — Requerimientos Clave

**Asignatura:** Probabilidad y Estadística · **Carrera:** ISC · **Entrega:** 27/05/2025

---

## Objetivo del sistema

Construir una aplicación que permita **seleccionar una distribución**, ingresar parámetros, generar datos aleatorios con tamaño de muestra configurable, calcular estadísticos y comparar resultados simulados vs. teóricos con visualización gráfica.

---

## Distribuciones a implementar (mínimo 6)

| Tipo               | Distribuciones requeridas                                 | Parámetros                    |
| ------------------ | --------------------------------------------------------- | ----------------------------- |
| **Discretas (≥3)** | Bernoulli, Binomial, Geométrica, Hipergeométrica, Poisson | `n`, `p`, `λ`, `N`, `K`, `n`  |
| **Continuas (≥3)** | Uniforme, Normal, Exponencial                             | `a`,`b` / `μ`,`σ` / `λ` o `β` |

**Validaciones obligatorias:** `0 < p ≤ 1`, `σ > 0`, `n > 0`, `a < b`.

---

## Funcionalidades mínimas del programa

- Selector de distribución con entrada dinámica de parámetros
- Generación de datos aleatorios (tamaño de muestra configurable)
- Cálculo de **media, varianza y desviación estándar** empírica
- Visualización gráfica:
  - Discretas → gráfica de barras (PMF)
  - Continuas → histograma normalizado + curva PDF teórica
- Tabla comparativa simulado vs. teórico por distribución
- Interpretación breve de resultados

### Fórmulas teóricas clave

| Distribución | Media teórica | Varianza teórica |
| ------------ | ------------- | ---------------- |
| Binomial     | \(np\)        | \(np(1-p)\)      |
| Poisson      | \(\lambda\)   | \(\lambda\)      |
| Geométrica   | \(1/p\)       | \((1-p)/p^2\)    |
| Normal       | \(\mu\)       | \(\sigma^2\)     |
| Uniforme     | \((a+b)/2\)   | \((b-a)^2/12\)   |
| Exponencial  | \(1/\lambda\) | \(1/\lambda^2\)  |

---

## Funcionalidades opcionales (bonus)

- Interfaz gráfica / app web/ tui
- Simulación de Ley de Grandes Números o TLC
- Comparación simultánea de dos distribuciones
- Exportación de resultados en CSV y gráficas como imagen
- README.md con instrucciones de instalación y uso

---

## Stack permitido

Python (`numpy`, `scipy`, `matplotlib`, `pandas`, `streamlit`), JavaScript (Chart.js, D3.js), MATLAB, R, Java, C/C++,GO

---

## Entregables

1. Código fuente completo (repositorio Git recomendado)
2. Reporte escrito en PDF con la siguiente estructura:
   - Portada → Introducción → Objetivo → Fundamento teórico
   - Descripción del programa → Resultados (gráficas + tablas)
   - Análisis e interpretación → Conclusiones → Referencias → Anexo
3. Capturas de pantalla de gráficas generadas
4. Ejemplos de ≥4 distribuciones implementadas
5. **Tabla comparativa por distribución:**

| Concepto            | Valor teórico | Valor simulado | Diferencia |
| ------------------- | ------------- | -------------- | ---------- |
| Media               |               |                |            |
| Varianza            |               |                |            |
| Desviación estándar |               |                |            |
| Tamaño de muestra   | —             |                | —          |

---

## Criterios de evaluación

| Criterio                                         | Puntos  |
| ------------------------------------------------ | ------- |
| Implementación ≥4 distribuciones                 | 15      |
| Generación correcta de datos simulados           | 10      |
| Cálculo de media, varianza y desviación estándar | 10      |
| Visualización gráfica clara                      | 10      |
| Comparación simulación vs. teoría                | 15      |
| Interpretación estadística                       | 15      |
| Reporte escrito y fundamento teórico             | 20      |
| Organización del código y documentación          | 5       |
| **Total**                                        | **100** |

---

## Preguntas guía para el reporte (interpretación)

- ¿Qué representan los parámetros de la distribución simulada?
- ¿La media/varianza simulada se aproxima a la teórica?
- ¿Qué ocurre cuando aumenta el tamaño de muestra?
- ¿Por qué los valores simulados nunca son exactamente iguales a los teóricos?
- ¿Qué diferencia gráfica hay entre distribuciones discretas y continuas?

---
