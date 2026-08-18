using Forge.Api.Accounts;
using Npgsql;

namespace Forge.Tests;

/// <summary>
/// paso-00 integration tests — they talk to the real docker-compose Postgres.
/// Run them with:  make up && make migrate && make test
///
/// If the DB is unreachable they SKIP (not fail), so <c>dotnet test</c> stays
/// green with Docker down. A later step upgrades these to hermetic Testcontainers.
/// </summary>
public class AccountStoreTests
{
    private static string ConnectionString =>
        Environment.GetEnvironmentVariable("FORGE_DB")
        ?? "Host=localhost;Port=5432;Database=forge;Username=forge;Password=forge";

    private static async Task<bool> DbReachableAsync()
    {
        try
        {
            await using var conn = new NpgsqlConnection(ConnectionString);
            await conn.OpenAsync();
            return true;
        }
        catch
        {
            return false;
        }
    }

    [SkippableFact]
    public async Task Creates_an_account_and_reads_it_back()
    {
        Skip.IfNot(await DbReachableAsync(), "Postgres not reachable — run `make up && make migrate`.");

        var store = new AccountStore(ConnectionString);

        var created = await store.CreateAsync("Alice");
        Assert.NotEqual(Guid.Empty, created.Id);
        Assert.Equal("Alice", created.Name);
        Assert.Equal(0m, created.Balance);

        var fetched = await store.GetAsync(created.Id);
        Assert.NotNull(fetched);
        Assert.Equal(created.Id, fetched!.Id);
    }

    [SkippableFact]
    public async Task Parameterized_query_neutralizes_sql_injection()
    {
        Skip.IfNot(await DbReachableAsync(), "Postgres not reachable — run `make up && make migrate`.");

        var store = new AccountStore(ConnectionString);

        // A classic injection payload. With parameters it is stored as a literal
        // name, never executed — so the accounts table survives.
        const string evil = "Robert'); DROP TABLE accounts;--";
        var created = await store.CreateAsync(evil);
        Assert.Equal(evil, created.Name);

        // The table is still there → we can still read from it.
        var fetched = await store.GetAsync(created.Id);
        Assert.NotNull(fetched);
    }
}
