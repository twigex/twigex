FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates tzdata media-types \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY twigex ./twigex
COPY frontend/dist ./frontend/dist
COPY templates ./templates
COPY LICENSE ./LICENSE
COPY i18n ./i18n

CMD ["./twigex", "start"]
