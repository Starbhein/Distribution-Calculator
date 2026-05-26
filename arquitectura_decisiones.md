# Registro de Decisiones Arquitectónicas (ADR)
## Simulador Avanzado de Distribuciones Estadísticas: Exprimiendo el Silicio

Este documento compila de forma exhaustiva las decisiones de ingeniería, diseño algorítmico y patrones de concurrencia adoptados en la construcción del simulador de distribuciones de probabilidad de alto rendimiento. Cada decisión está fundamentada en tres pilares inquebrantables: **máxima estabilidad numérica, rendimiento bruto cercano al hardware y desacoplamiento modular riguroso.**

---

## 1. Filosofía de Diseño: "Hasta la última gota de silicio"

El objetivo del proyecto trasciende la creación de una simple calculadora escolar; se enfoca en el diseño de un sistema de sistemas capaz de simular millones de eventos estocásticos en fracciones de segundo sin degradar la experiencia de usuario (UX) en la terminal. Para lograrlo, se rechazó cualquier abstracción innecesaria, se evitó la contención de memoria a nivel de hardware y se diseñó una arquitectura de flujo unidireccional reactiva donde el cómputo pesado corre de forma asíncrona, dejando la interfaz de usuario libre para renderizar a un marco de tasa constante (60 FPS).

---

## 2. Núcleo Matemático y Estabilidad Numérica

### Contexto
La simulación de distribuciones como Poisson o Binomial tradicionalmente se enseña mediante fuerza bruta (Simulación Literal con bucles de tiempo $\mathcal{O}(N)$) o mediante algoritmos clásicos como el método de Knuth ($P \leftarrow \prod U_i$). 

### Decisiones y Justificación
1. **Adopción de la Transformada Inversa mediante Relaciones de Recurrencia:** Se descartó la simulación de fuerza bruta debido a su costo computacional prohibitivo en parámetros grandes (ej. $N = 10,000$). En su lugar, se implementó el método analítico de la **Transformada Inversa**, generando un único número aleatorio uniforme $U \sim (0,1)$ y acumulando la CDF al vuelo. Para evitar el cálculo costoso de factoriales y combinatorias que provocarían desbordamientos (*overflow*), se adoptaron **relaciones de recurrencia**, donde la probabilidad del paso $k$ se calcula a partir del paso $k-1$ mediante multiplicaciones puras:
   $$	ext{Poisson: } P(k) = P(k-1) \cdot rac{\lambda}{k}$$
   Esto reduce la complejidad temporal de $\mathcal{O}(N)$ a proporcional a la media computacional de la distribución ($\mathcal{O}(\lambda)$ o $\mathcal{O}(Np)$).
2. **Control Extremo de Underflow (El Patrón Guardián):** El uso de `math.Exp(-lambda)` para el término inicial de Poisson ($k=0$) sufre de *underflow* numérico en el estándar IEEE 754 (`float64`) cuando $\lambda > 745$, degradando el valor a `0.0` puro y provocando bucles infinitos en el generador. En lugar de sobrecargar el núcleo con espacio logarítmico complejo para casos típicos de TUI, se delegó la validación estricta a la capa tipada intermedia e interfaz de usuario, topando de forma segura los parámetros numéricos de entrada.

---

## 3. Concurrencia Extrema: Patrón de Confinamiento y Bloques Contiguos

### Contexto
Para llenar un búfer masivo de datos simulados (ej. $1,000,000$ de muestras), la ejecución secuencial desaprovecha las capacidades multihilo de los procesadores modernos. Sin embargo, una mala estrategia de paralelismo introduce contención por cerrojos (*mutex*) o degradación del rendimiento por la arquitectura de caché del CPU.

### Decisiones y Justificación
1. **Patrón de Confinamiento Estricto:** Se prohibió compartir variables o canales comunes dentro de los bucles de generación estocástica. El búfer principal de memoria se pre-asigna una sola vez como un *slice* contiguo en memoria. Este búfer se subdivide en $4$ bloques disjuntos y contiguos (uno por cada *goroutine* de procesamiento). Cada *goroutine* es dueña absoluta de su segmento de memoria, confinándose el acceso a su espacio asignado.
2. **Aislamiento de Estado de Generadores Aleatorios (PRNG):** La librería estándar de Go tradicionalmente comparte un *mutex* global en el generador aleatorio. Para romper este cuello de botella, cada *goroutine* instancia su propio generador local basado en el paquete moderno `math/rand/v2` utilizando el algoritmo determinista **PCG (Permuted Congruential Generator)**. Cada hilo recibe una semilla única y local, garantizando generación pseudoaleatoria paralela pura a la velocidad nativa del silicio.
3. **Mitigación Absoluta de False Sharing:** Al estructurar los bloques de memoria de forma **contigua y no intercalada** (ej. Hilo 1 escribe de `0` a `N/4`, Hilo 2 de `N/4` a `N/2`), se garantiza que las líneas de caché del procesador (L1/L2, usualmente de 64 bytes) sean leídas y escritas por un único núcleo a la vez. Esto elimina el fenómeno de *False Sharing* (falsa compartición), permitiendo una escalabilidad casi lineal en relación al número de núcleos físicos del CPU.

---

## 4. Streaming Analytics: Algoritmo de Welford Paralelo y Reducción

### Contexto
El cálculo tradicional de la media y la varianza empírica requiere dos pasadas sobre el búfer de memoria (`AnalyzeBuffer` secuencial): una para encontrar la media y otra para acumular los cuadrados de las diferencias. Esto duplica el tráfico de lectura en la memoria RAM, castigando el rendimiento.

### Decisiones y Justificación
1. **Cómputo en una Sola Pasada (Online Welford):** Se integró el **Algoritmo de Welford** directamente dentro del bucle de generación de números aleatorios de cada *goroutine*. Conforme la transformada inversa arroja un número simulado, este se procesa inmediatamente a través del acumulador local de Welford actualizando la cuenta, la media en ejecución y la suma de cuadrados de diferencias métricas ($M_2$) sin perder precisión decimal frente a la cancelación catastrófica de coma flotante.
2. **Reducción Cohesiva (Parallel Merge):** Al terminar la ejecución paralela mediante un `sync.WaitGroup`, el hilo principal no vuelve a leer el búfer gigante de memoria. En su lugar, realiza un proceso de reducción matemática combinando los 4 acumuladores locales utilizando las fórmulas de agregación estadística de varianzas:
   $$N_{	ext{comb}} = N_A + N_B$$
   $$\mu_{	ext{comb}} = \mu_A + \delta \cdot rac{N_B}{N_{	ext{comb}}}$$
   $$M_{2,	ext{comb}} = M_{2,A} + M_{2,B} + \delta^2 \cdot rac{N_A \cdot N_B}{N_{	ext{comb}}}$$
   Donde $\delta = \mu_B - \mu_A$. Esto permite consolidar métricas globales exactas sobre millones de datos simulados en tiempo constante $\mathcal{O}(1)$ tras la sincronización de hilos.

---

## 5. Arquitectura del API: Pureza de Funciones y Desacoplamiento (DTO)

### Contexto
El acoplamiento directo entre las estructuras teóricas matemáticas y el motor de simulación empírico limita la escalabilidad del sistema y rompe el principio de responsabilidad única.

### Decisiones y Justificación
1. **Funciones Puras en el Motor:** El `SimulatorEngine` fue diseñado para ser completamente ciego e ignorante de la existencia de estructuras teóricas de probabilidad. Sus métodos de generación (ej. `FillPoisson`) solo consumen tipos primitivos crudos del lenguaje (`[]float64`, `float64`), lo que maximiza la velocidad de ejecución y simplifica radicalmente los tests unitarios analíticos basados en escenarios estáticos con semillas fijas.
2. **El Patrón de Cápsulas de Trabajo (Wrapper / DTO):** Para estandarizar las órdenes de trabajo que viajan desde la interfaz visual hacia el motor de cómputo, se implementó la interfaz `EmpiricalSimulator` basada en un único contrato rígido: `FillBuffer(buffer []float64) error`. Cada distribución implementa su propia cápsula o *struct* de trabajo (ej. `PoissonTask`). Este objeto encapsula los parámetros específicos capturados por el usuario (ej. `Lambda`), abstrayendo las firmas complejas para que el *worker* secundario pueda ejecutar cualquier simulación de manera polimórfica y ciega aplicando los principios SOLID (Open/Closed Principle).

---

## 6. Interfaz de Usuario Reactiva: Bubble Tea V2 y Diseño Dashboard

### Contexto
Renderizar interfaces complejas en la terminal (TUI) suele resultar en arquitecturas espagueti debido a la gestión desordenada del estado del teclado y el redibujado de la pantalla.

### Decisiones y Justificación
1. **Paradigma Arquitectónico Elm:** Se adoptó estrictamente el flujo unidireccional de datos nativo de **Bubble Tea V2** (Modelo, Vista, Actualización). El estado muta única y exclusivamente a través de la recepción de Mensajes independientes (`tea.Msg`), blindando la consistencia de los datos.
2. **Estructura del Enrutador por Sub-modelos:** El modelo principal (`MainModel`) actúa exclusivamente como un director de orquesta y una Máquina de Estados Finitos (`stateMenu`, `stateForm`, `stateLoading`, `stateResults`). Los flujos de interacción complejos están delegados a sub-modelos completamente independientes y encapsulados (`MenuModel`, `FormModel`).
3. **Comunicación mediante Mensajes de Transición (Gritos al Vacío):** Los sub-modelos jamás alteran estados de otros módulos. Cuando el menú detecta un *Enter*, emite un comando con el mensaje `MsgSelectedDistribution`. El Enrutador principal intercepta este mensaje, apaga el menú, invoca dinámicamente la fábrica de cajas de texto del formulario (`BuildInputs`) e inyecta el nuevo estado visual de forma limpia.
4. **Validación Bidireccional Adaptativa de UX:** El proceso de traducción de datos desde la captura de texto (`string`) hacia el procesamiento matemático (`float64`) se realiza mediante un `Parser` intermedio seguro con `strconv.ParseFloat`. Si ocurre un error de entrada, el Enrutador no bloquea la aplicación; despacha un mensaje de error específico apuntando al índice del fallo. El `FormModel` reacciona pintando en tiempo real un cartel en rojo vibrante (`#FC869C`) bajo el campo exacto del error, el cual se limpia automáticamente de forma dinámica en cuanto el usuario presiona una nueva tecla para corregir su entrada.
5. **Layout Dashboard Segmentado con Lipgloss:** Utilizando las nuevas características de vistas fluidas de Bubble Tea V2 (`tea.View`), la aplicación captura de forma dinámica los cambios de tamaño de la ventana de la terminal (`tea.WindowSizeMsg`). Al transicionar al tablero, la pantalla se maqueta dividiendo el espacio total en un diseño responsivo de tarjetas independientes mediante `lipgloss.JoinHorizontal`:
   * **Panel de Control (30% de ancho):** Tarjeta con borde redondeado dedicada en exclusiva a los inputs dinámicos del formulario y al estatus interactivo del botón de envío.
   * **Panel de Visualización (70% de ancho):** Espacio dedicado al renderizado de tablas informativas omniscientes y componentes asíncronos distribuidos. Durante el procesamiento masivo, este panel renderiza un componente gráfico animado (`spinner.Model`) que corre libremente a 60 FPS gracias a que las 4 *goroutines* matemáticas fueron despachadas fuera del bucle de eventos mediante comandos no bloqueantes de la TUI (`tea.Cmd`).

---

## 7. Conclusión

La combinación simbiótica de estas decisiones arquitectónicas ha dado como resultado un software de sistemas ejemplar para la terminal. La rigidez y pureza matemática de las operaciones de reducción paralela garantizan que cada ciclo de reloj del procesador se traduzca en progreso directo de simulación, mientras que el desacoplamiento estricto de componentes basados en eventos asegura que el mantenimiento y la extensión de nuevas distribuciones sea una tarea trivial, segura y limpia.
