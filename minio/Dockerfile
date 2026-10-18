# Stage 1: Use the MinIO Client to build configuration files
FROM minio/mc:RELEASE.2026-01-16T23-45-56Z AS config-builder

# Set up local workspace variables
WORKDIR /config

# Create your custom startup script
RUN echo '#!/bin/sh' > entrypoint.sh && \
    echo 'minio server /data --console-address ":9001" &' >> entrypoint.sh && \
    echo 'PID=$!' >> entrypoint.sh && \
    echo 'until (/usr/bin/mc alias set local http://localhost:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD"); do sleep 1; done' >> entrypoint.sh && \
    echo '/usr/bin/mc mb --ignore-existing local/raw-images' >> entrypoint.sh && \
    echo '/usr/bin/mc admin user add local "$GO_USER" "$GO_PASSWORD"' >> entrypoint.sh && \
    echo '/usr/bin/mc admin policy attach local writeonly --user "$GO_USER"' >> entrypoint.sh && \
    echo '/usr/bin/mc admin user add local "$PYTHON_USER" "$PYTHON_PASSWORD"' >> entrypoint.sh && \
    echo '/usr/bin/mc admin policy attach local readonly --user "$PYTHON_USER"' >> entrypoint.sh && \
    echo 'wait $PID' >> entrypoint.sh && \
    chmod +x entrypoint.sh

# Stage 2: Final Production Ready Image
FROM minio/minio:RELEASE.2026-01-16T22-48-28Z

# Copy the entrypoint script from the builder stage
COPY --from=config-builder /config/entrypoint.sh /usr/bin/entrypoint.sh

# Set default production-safe fallbacks (Overridden by your real Env Variables)
ENV MINIO_ROOT_USER=admin-super-user
ENV MINIO_ROOT_PASSWORD=super-secret-admin-password
ENV GO_USER=go_downloader_user
ENV GO_PASSWORD=go_password_123
ENV PYTHON_USER=python_worker_user
ENV PYTHON_PASSWORD=python_password_123

# Execute your wrapped launch command
ENTRYPOINT ["/usr/bin/entrypoint.sh"]
