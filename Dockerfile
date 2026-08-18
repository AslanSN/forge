FROM mcr.microsoft.com/dotnet/sdk:10.0 AS InitBuild
WORKDIR /src
COPY src/Forge.Api/Forge.Api.csproj  ./src/Forge.Api/
RUN dotnet restore src/Forge.Api/Forge.Api.csproj
COPY src/Forge.Api ./src/Forge.Api/
RUN dotnet publish src/Forge.Api -o /app/out


FROM mcr.microsoft.com/dotnet/aspnet:10.0
WORKDIR /app
COPY --from=InitBuild /app/out .
RUN useradd -r -s /bin/false nonroot
ENV FORGE_URLS=http://0.0.0.0:8080
USER nonroot
EXPOSE 8080
ENTRYPOINT ["dotnet", "/app/Forge.Api.dll"]
