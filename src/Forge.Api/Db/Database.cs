namespace Forge.Api.Db;

/// <summary>
/// Resolves the Postgres connection string.
/// paso-00: read it from the environment so the same binary talks to the
/// docker-compose DB locally and to a real DB in CI/prod — never hard-coded.
/// </summary>
public static class Database
{
    /// <summary>Dev default: matches <c>docker-compose.yml</c>.</summary>
    public const string DevConnectionString =
        "Host=localhost;Port=5432;Database=forge;Username=forge;Password=forge";

    public static string ConnectionString =>
        Environment.GetEnvironmentVariable("FORGE_DB") is { Length: > 0 } cs
            ? cs
            : DevConnectionString;
}
