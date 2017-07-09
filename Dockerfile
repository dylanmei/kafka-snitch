FROM scratch

COPY bin/release/kafka-snitch /kafka-snitch
ENTRYPOINT ["/kafka-snitch"]
