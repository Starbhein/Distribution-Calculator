# Distribution Calculator

> [!NOTE]
> Operations are subjected to a measurement error Epsilon=1^-12

```Golang

const epsilonFailure = 1e-12

```

## Discrete Distribution Variable **Distributions covered**

- Binomial Distribution `Under development`
  The probability of exactly k successes when k it's an integer

```math
P(x = k) = C_k^n \, p^k q^{n-k} = \frac{n!}{k!(n-k)!} \, p^k q^{n-k}
```

- Poisson Distribution `Under development`
- Geometric Distribution `Under development`
- Hypergeometric Distribution `Under development`

## Continuous Distribution Variable **Distribution covered**

- Regular Distribution `Under development`
- Uniform Distribution `Under development`
- Exponential Distribution `Under development`

### Main Functionalities

> [!NOTE]
> This capabilities are implemented for every distribution.

1. Charts
   Each chart includes the following elements that can be rendered in the tui or exported.
2. Theoretical & empirical average.
3. Theoretical & empirical variance.
4. Theoretical & empirical Standard deviation.
5. Brief explanation

#### How the program works

##### **Double buffered concurrency pattern with use of containment**

##### Factorial optimization via the Gamma function

#####
