FROM bitnami/minideb:stretch
RUN install_packages ca-certificates
ADD build/km-linux-x64 /app/km

CMD ["/app/km"]