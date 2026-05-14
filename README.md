# Distribution Calculator

> [!NOTE]
> The operations has a measurement error Epsilon=1^-12

```Golang

const epsilonFailure = 1e-12

```

## Discrete Distribution Variable **Distribution covered**

- Binomial Distribution `On development`
  Probability on a k case where k it's a real int number

```math
P(x = k) = C_k^n \, p^k q^{n-k} = \frac{n!}{k!(n-k)!} \, p^k q^{n-k}
```

- Poisson Distribution `On development`
- Geometric Distribution `On development`
- Hypergeometric Distribution `On development`

## Continuous Distribution Variable **Distribution covered**

- Regular Distribution `On development`
- Uniform Distribution `On development`
- Exponential Distribution `On development`

### Main Functionalities

> [!NOTE]
> All this capabilities are designed and implemented for each different Distribution.

1. Charts
   Every chart has the next elements and could be viewed on the tui or exported.
2. Theoretical & empirical average.
3. Theoretical & empirical variance.
4. Theoretical & empirical Standar deviation.
5. Briefly explanation

#### How the program works

##### **Double buffered concurrency pattern with use of containment**

##### n! improvement, use of Gamma function

#####
