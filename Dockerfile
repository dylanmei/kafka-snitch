FROM gcr.io/distroless/base
COPY bin/release/kafka-snitch /kafka-snitch
ENTRYPOINT ["/kafka-snitch"]
