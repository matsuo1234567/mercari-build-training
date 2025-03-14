FROM alpine

# Install Go 1.22.4
RUN apk add --no-cache wget tar

# Download and install Go 1.22.4 for Linux
RUN wget https://go.dev/dl/go1.22.4.linux-arm64.tar.gz && \
    tar -C /usr/local -xzf go1.22.4.linux-arm64.tar.gz && \
    rm go1.22.4.linux-arm64.tar.gz

# Add Go binary directory to PATH
ENV PATH=$PATH:/usr/local/go/bin

# Create Users and Groups
RUN addgroup -S mercari && adduser -S trainee -G mercari
# RUN chown -R trainee:mercari /path/to/db

USER trainee

CMD ["go", "version"]
