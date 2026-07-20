using Npgsql;

namespace Forge.Api.Accounts;

/// <summary>
/// paso-00 · the database, by hand. No ORM — raw SQL over Npgsql, ALWAYS
/// parameterized (never string-interpolated; see docs/paso-00-la-bd-a-pelo.md).
///
/// NOTE: <c>balance</c> is a mutable column here. That is deliberately naive.
/// paso-02 (the double-spend race) will break it and force an append-only
/// <c>entries</c> table where the balance is derived, not stored.
/// </summary>
public sealed class AccountStore(string connectionString)
{
    public async Task<AccountResponse> CreateAsync(string name, CancellationToken ct = default)
    {
        const string sql =
            """
            INSERT INTO accounts (name)
            VALUES (@name)
            RETURNING id, name, balance;
            """;

        await using var conn = new NpgsqlConnection(connectionString);
        await conn.OpenAsync(ct);
        await using var cmd = new NpgsqlCommand(sql, conn);
        cmd.Parameters.AddWithValue("name", name); // a parameter, NOT string concatenation
        await using var reader = await cmd.ExecuteReaderAsync(ct);
        await reader.ReadAsync(ct);
        return Map(reader);
    }

    public async Task<AccountResponse?> GetAsync(Guid id, CancellationToken ct = default)
    {
        const string sql =
            """
            SELECT id, name, balance
            FROM accounts
            WHERE id = @id;
            """;

        await using var conn = new NpgsqlConnection(connectionString);
        await conn.OpenAsync(ct);
        await using var cmd = new NpgsqlCommand(sql, conn);
        cmd.Parameters.AddWithValue("id", id);
        await using var reader = await cmd.ExecuteReaderAsync(ct);
        return await reader.ReadAsync(ct) ? Map(reader) : null;
    }

    private static AccountResponse Map(NpgsqlDataReader r) =>
        new(r.GetGuid(0), r.GetString(1), r.GetDecimal(2));
}
