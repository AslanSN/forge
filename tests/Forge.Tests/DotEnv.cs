using System.Runtime.CompilerServices;

namespace Forge.Tests;

/// <summary>
/// paso-00-level scaffolding (written for you). Loads the repo's <c>.env</c> into
/// the process environment before any test runs.
/// </summary>
/// <remarks>
/// <para>
/// The Makefile sources <c>.env</c> before <c>dotnet test</c> (see <c>make test</c>).
/// An IDE does not: Rider launches the test host straight from the build output,
/// so <c>FORGE_DB</c> is unset and <c>ConnectionString</c> falls back to port 5432
/// — which on this machine is a *different* project's Postgres. The connection
/// fails, <c>Skip.IfNot</c> fires, and the integration tests report as skipped.
/// </para>
/// <para>
/// A skip is the failure mode that hurts most here, because it looks like success:
/// the DB half of paso-01 appears done without ever having run. Same trap the Go
/// line had. This makes the IDE and the terminal see the same environment.
/// </para>
/// <para>
/// Real environment variables always win — a value already set is never
/// overwritten, so <c>make test</c> and CI keep control.
/// </para>
/// </remarks>
internal static class DotEnv
{
    [ModuleInitializer]
    internal static void Load()
    {
        if (FindEnvFile() is not { } path) return;

        foreach (var raw in File.ReadLines(path))
        {
            var line = raw.Trim();
            if (line.Length == 0 || line[0] == '#') continue;
            if (line.StartsWith("export ", StringComparison.Ordinal)) line = line[7..].TrimStart();

            // Split on the FIRST '=' only: FORGE_DB's value is a connection
            // string and is full of them.
            var eq = line.IndexOf('=');
            if (eq <= 0) continue;

            var key = line[..eq].TrimEnd();
            var value = line[(eq + 1)..].Trim();
            if (value.Length >= 2 && (value[0] == '"' || value[0] == '\'') && value[^1] == value[0])
                value = value[1..^1];

            if (Environment.GetEnvironmentVariable(key) is null)
                Environment.SetEnvironmentVariable(key, value);
        }
    }

    /// <summary>
    /// Walks up from the test assembly's output directory looking for .env, and
    /// stops at the repo root so it can never pick up a stray .env from outside
    /// the checkout. Returns null when there is none — a fresh clone and CI have
    /// no .env, and that is a valid setup, not an error.
    /// </summary>
    private static string? FindEnvFile()
    {
        for (var dir = new DirectoryInfo(AppContext.BaseDirectory); dir is not null; dir = dir.Parent)
        {
            var candidate = Path.Combine(dir.FullName, ".env");
            if (File.Exists(candidate)) return candidate;

            // .git is a directory in a normal clone and a file in a worktree.
            if (Path.Exists(Path.Combine(dir.FullName, ".git"))) break;
        }
        return null;
    }
}
