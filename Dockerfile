FROM alpine:3.22
WORKDIR /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup

COPY build/paye-ton-kawa--products /app/paye-ton-kawa--products

RUN chown appuser:appgroup /app/paye-ton-kawa--products && chmod +x /app/paye-ton-kawa--products

EXPOSE 8083

USER appuser

ENTRYPOINT ["/app/paye-ton-kawa--products"]
