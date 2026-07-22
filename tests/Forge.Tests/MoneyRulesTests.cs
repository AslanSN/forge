using Forge.Api.Accounts;

namespace Forge.Tests;

public class MoneyRulesTests
{
    // Illustrative — WHY money is decimal, not float. This one PASSES already;
    // run it and see the drift with your own eyes.
    [Fact]
    public void Float_drifts_but_decimal_is_exact()
    {
        double f = 0.0;
        for (var i = 0; i < 10; i++) f += 0.1;
        Assert.NotEqual(1.0, f); // 0.1 added ten times is NOT 1.0 in binary float

        decimal d = 0m;
        for (var i = 0; i < 10; i++) d += 0.1m;
        Assert.Equal(1.0m, d);   // decimal is exact
    }

    // YOUR TURN — implement MoneyRules.NormalizeAmount to make these green.
    [Fact]
    public void Accepts_a_normal_amount()
    {
        var (amount, error) = MoneyRules.NormalizeAmount(12.34m);
        Assert.Null(error);
        Assert.Equal(12.34m, amount);
    }

    [Theory]
    [InlineData(0)]
    [InlineData(-5)]
    public void Rejects_non_positive_amounts(int raw)
    {
        var (_, error) = MoneyRules.NormalizeAmount(raw);
        Assert.NotNull(error);
    }

    [Fact]
    public void Rejects_more_than_two_decimals()
    {
        var (_, error) = MoneyRules.NormalizeAmount(1.234m);
        Assert.NotNull(error);
    }
}
