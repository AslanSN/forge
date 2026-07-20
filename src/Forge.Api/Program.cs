using Forge.Api.Accounts;
using Forge.Api.Db;

var builder = WebApplication.CreateBuilder(args);

// Deterministic local URL so the README's curl examples are always correct.
builder.WebHost.UseUrls(Environment.GetEnvironmentVariable("FORGE_URLS") ?? "http://localhost:5000");

// paso-00: a single raw-SQL store, no ORM. Registered as a singleton because it
// holds only a connection string; Npgsql pools the underlying connections.
builder.Services.AddSingleton(new AccountStore(Database.ConnectionString));

var app = builder.Build();

app.MapGet("/", () => "forge · paso-00 — the database, by hand. POST /accounts, GET /accounts/{id}.");

app.MapPost("/accounts", async (CreateAccountRequest req, AccountStore store, CancellationToken ct) =>
{
    var (name, error) = AccountRules.NormalizeName(req.Name);
    if (error is not null)
        return Results.BadRequest(new { error });

    var account = await store.CreateAsync(name!, ct);
    return Results.Created($"/accounts/{account.Id}", account);
});

app.MapGet("/accounts/{id:guid}", async (Guid id, AccountStore store, CancellationToken ct) =>
{
    var account = await store.GetAsync(id, ct);
    return account is null ? Results.NotFound() : Results.Ok(account);
});

app.Run();

// Exposed so later pasos can drive the app with WebApplicationFactory in tests.
public partial class Program;
