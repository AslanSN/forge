namespace Forge.Api.Accounts;

public sealed record CreateAccountRequest(string? Name);

public sealed record AccountResponse(Guid Id, string Name, decimal Balance);

/// <summary>
/// Account input rules as PURE logic, so they are unit-testable without a database.
/// (Keeping validation out of the I/O path is the point — see AccountRulesTests.)
/// </summary>
public static class AccountRules
{
    public const int MaxNameLength = 100;

    /// <summary>Returns the normalized name, or an error message. No I/O.</summary>
    public static (string? Name, string? Error) NormalizeName(string? raw)
    {
        var name = raw?.Trim();
        if (string.IsNullOrEmpty(name))
            return (null, "name is required");
        if (name.Length > MaxNameLength)
            return (null, $"name must be at most {MaxNameLength} characters");
        return (name, null);
    }
}

/// <summary>
/// paso-01 · money rules. YOUR TURN — implement NormalizeAmount so MoneyRulesTests pass.
/// See docs/paso-01-money.md. PURE (no I/O) so it is unit-testable without a database.
/// </summary>
public static class MoneyRules
{
    public const int MaxDecimalScale = 2;
    /// <summary>
    /// Validate a monetary amount. Return the amount, or an error message.
    /// Rules: must be > 0, and at most 2 decimal places.
    /// </summary>
    public static (decimal Amount, string? Error) NormalizeAmount(decimal raw)
    {
        if (raw <= 0)
            return (raw, "value must be positive");

        if (raw.Scale > MaxDecimalScale)
            return (raw, $"value cannot have more than {MaxDecimalScale} decimals");

        return (raw, null);
    }

}
