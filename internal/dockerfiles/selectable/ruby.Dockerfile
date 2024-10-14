# Use Ruby version as an argument
# WIP: base image
ARG rubyVersion
FROM docker.io/library/ruby:${rubyVersion}
LABEL com.zeabur.image-type="containerized"

# Install system dependencies and Node.js, npm, yarn, pnpm
RUN apt-get update -qq && apt-get install -y \
    postgresql-client \
    nodejs \
    npm && \
    npm install -g yarn pnpm

# Set the working directory
WORKDIR /myapp

# Copy application source code
COPY . /myapp

# Install Ruby dependencies
RUN bundle install

# Argument for package manager (yarn, pnpm, npm or "" (none))
ARG nodePackageManager

# Install Node.js dependencies based on the package manager
RUN if [ "$nodePackageManager" = "yarn" ]; then \
      yarn install; \
    elif [ "$nodePackageManager" = "pnpm" ]; then \
		pnpm install; \
    elif [ "$nodePackageManager" = "npm" ]; then \
		npm install; \
    fi

# Build command (optional)
ARG build
RUN if [ -n "$build" ]; then $build; fi

# Start command
ARG start
CMD $start
