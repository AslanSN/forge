using Forge.Api.Accounts;
using Npgsql;

namespace Forge.Tests;

/// <summary>
/// paso-01 · deposit/withdraw against the real DB. YOUR TURN — implement the store
/// methods to make these green. They skip when Postgres is down (run `make up`).
/// </summary>
public class AccountMoneyTests
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
    public async Task Deposit_then_withdraw_keeps_an_exact_balance()
    {
        Skip.IfNot(await DbReachableAsync(), "Postgres not reachable — run `make up && make migrate`.");
        var store = new AccountStore(ConnectionString);
        var acc = await store.CreateAsync("Money Alice");

        var (afterDeposit, depErr) = await store.DepositAsync(acc.Id, 100.00m);
        Assert.Null(depErr);
        Assert.Equal(100.00m, afterDeposit!.Balance);

        var (afterWithdraw, wErr) = await store.WithdrawAsync(acc.Id, 30.50m);
        Assert.Null(wErr);
        Assert.Equal(69.50m, afterWithdraw!.Balance);
    }

    [SkippableFact]
    public async Task Withdraw_more_than_balance_is_rejected()
    {
        Skip.IfNot(await DbReachableAsync(), "Postgres not reachable — run `make up && make migrate`.");
        var store = new AccountStore(ConnectionString);
        var acc = await store.CreateAsync("Broke Bob");
        await store.DepositAsync(acc.Id, 10.00m);

        var (_, error) = await store.WithdrawAsync(acc.Id, 999.00m);
        Assert.NotNull(error); // overdraft rejected

        var reread = await store.GetAsync(acc.Id);
        Assert.Equal(10.00m, reread!.Balance); // balance unchanged
    }
}
