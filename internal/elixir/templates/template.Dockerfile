# Base image
FROM elixir:{{.ElixirVer}} AS build

# Set environment variables
ENV MIX_ENV=prod
ENV PORT=8080

# Install Hex package manager and Rebar
RUN mix local.hex --force && \
    mix local.rebar --force

# Set working directory
WORKDIR /app

# Copy config files
COPY config/* /app/config/

# Copy the mix.exs and mix.lock files
COPY mix.exs /app
COPY mix.lock /app

# Fetch and compile production dependencies
RUN mix deps.get --only prod

# Copy the project files to the working directory
COPY . /app

# Compile the project
RUN mix compile

{{ if .ElixirEcto }}
# Create Ecto database
RUN mix ecto.create

# Run database migrations
RUN mix ecto.migrate

# Run seeds
RUN mix run priv/repo/seeds.exs
{{ else }}
# Deploy assets (if applicable)
RUN mix assets.deploy
{{ end }}

{{ if .ElixirPhoenix }}
# Start the Phoenix server
CMD mix phx.server
{{ else }}
CMD mix run --no-halt
{{ end }}
