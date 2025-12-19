FROM mcr.microsoft.com/devcontainers/go:1-1.22

RUN apt-get update \
	&& apt-get install -y --no-install-recommends make ca-certificates curl python3-pip \
	&& rm -rf /var/lib/apt/lists/*

# Install golangci-lint (keeps in sync with pre-commit config) and pre-commit.
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
	| sh -s -- -b /usr/local/bin v2.1.0 \
	&& pip3 install --no-cache-dir pre-commit
