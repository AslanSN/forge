using Forge.Api.Accounts;

namespace Forge.Tests;

/// <summary>Pure unit tests — no database, always run (green even with Docker down).</summary>
public class AccountRulesTests
{
    [Fact]
    public void Trims_and_accepts_a_valid_name()
    {
        var (name, error) = AccountRules.NormalizeName("  Alice  ");
        Assert.Null(error);
        Assert.Equal("Alice", name);
    }

    [Theory]
    [InlineData(null)]
    [InlineData("")]
    [InlineData("   ")]
    public void Rejects_empty_names(string? raw)
    {
        var (name, error) = AccountRules.NormalizeName(raw);
        Assert.Null(name);
        Assert.NotNull(error);
    }

    [Fact]
    public void Rejects_overlong_names()
    {
        var (name, error) = AccountRules.NormalizeName(new string('x', AccountRules.MaxNameLength + 1));
        Assert.Null(name);
        Assert.NotNull(error);
    }
}
