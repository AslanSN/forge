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
