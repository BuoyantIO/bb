FROM gcr.io/runconduit/base:2017-10-30.01
RUN apt-get update
RUN apt-get install -y ca-certificates
RUN mkdir /app
ADD target/bb /app/
ENTRYPOINT ["/app/bb"]