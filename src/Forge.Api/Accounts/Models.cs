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
