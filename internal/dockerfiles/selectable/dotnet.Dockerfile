ARG dotnetVersion
ARG submoduleDir

FROM mcr.microsoft.com/dotnet/sdk:${dotnetVersion} AS build
WORKDIR /source

COPY *.csproj* ./
RUN dotnet restore

# copy everything else and build app
COPY . ./
WORKDIR /source
RUN dotnet publish -c release -o /app

FROM scratch AS static-serverless
LABEL com.zeabur.image-type="static"
COPY --from=build /app/wwwroot /

FROM zeabur/zbpack-static-containerized AS static-containerized
LABEL com.zeabur.image-type="containerized"
COPY --from=build /app/wwwroot .

FROM mcr.microsoft.com/dotnet/aspnet:${dotnetVersion} AS runtime
LABEL com.zeabur.image-type="containerized"
ARG out
ENV PORT=8080
WORKDIR /app
COPY --from=build /app ./
CMD ASPNETCORE_URLS=http://+:$PORT dotnet ${out}.dll
