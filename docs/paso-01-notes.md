## PostgreSQL
### Numeric
- Pros:
  1. Very large number of digits.
  2. Recommended for storing monetary amounts (exactness requiered)
  3. Calculations with `numeric` yield exact results where possible.
- Cons:
  1. Calculations very slow compared to integer types. 

#### Precision
 The total count of significant digits in the whole number = the number of digits to both sides of the decimal point.
#### Scale
Count of decimal digits in the fractional part, to the right of the decimal point.

<blockquote>
23.5141 = precision of 6 + scale of 4
</blockquote>

Integers = scale of 0.

#### Special values
1. Infinity
2. -Infinity
3. NaN

### Floating-Point Types (real & double precision)

- Cons:
  1. They are inexact variabe precision numeric types.
  2. Real has 6 points of precition while double precision of at last 15. But values too large or too small may generate an error.
  3. May round if number is too high. 
  4. Numbers too close to zero that are not representable as distinct from zero will cause an underflow error.

## .NET's answer
```csharp
public readonly struct Decimal : IComparable<decimal>, IConvertible, IEquatable<decimal>, IParsable<decimal>, ISpanParsable<decimal>, IUtf8SpanParsable<decimal>, System.Numerics.IAdditionOperators<decimal,decimal,decimal>, System.Numerics.IAdditiveIdentity<decimal,decimal>, System.Numerics.IComparisonOperators<decimal,decimal,bool>, System.Numerics.IDecrementOperators<decimal>, System.Numerics.IDivisionOperators<decimal,decimal,decimal>, System.Numerics.IEqualityOperators<decimal,decimal,bool>, System.Numerics.IFloatingPoint<decimal>, System.Numerics.IFloatingPointConstants<decimal>, System.Numerics.IIncrementOperators<decimal>, System.Numerics.IMinMaxValue<decimal>, System.Numerics.IModulusOperators<decimal,decimal,decimal>, System.Numerics.IMultiplicativeIdentity<decimal,decimal>, System.Numerics.IMultiplyOperators<decimal,decimal,decimal>, System.Numerics.INumber<decimal>, System.Numerics.INumberBase<decimal>, System.Numerics.ISignedNumber<decimal>, System.Numerics.ISubtractionOperators<decimal,decimal,decimal>, System.Numerics.IUnaryNegationOperators<decimal,decimal>, System.Numerics.IUnaryPlusOperators<decimal,decimal>, System.Runtime.Serialization.IDeserializationCallback, System.Runtime.Serialization.ISerializable
```

decimal es limitado pero exacto mientras que double / float64 es muy amplio pero inexacto. 0.1 no se representa exactamente en binario. 

## Go's answer
Go elimina el problema, no tiene un tipo decimal todos sus números son enteros.
Se piensa en centavos (int64) o en la versión monteria menor que se esté manejando.
Desaparece problema de precisión o escala porque **no hay fracciones binarias involucradas**
