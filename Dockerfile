FROM python:3.9-slim AS builder

RUN apt-get update && apt-get install -y \
    gcc g++ make golang default-jdk curl \
    libomp-dev binutils build-essential \
    && rm -rf /var/lib/apt/lists/*

RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
RUN rustup default stable && rustup target add x86_64-unknown-linux-gnu

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin"

ENV JAVA_HOME=/usr/lib/jvm/default-java
ENV PATH=$JAVA_HOME/bin:$PATH

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM python:3.9-slim

COPY --from=builder /usr/bin/gcc /usr/bin/
COPY --from=builder /usr/bin/g++ /usr/bin/
COPY --from=builder /usr/local/go /usr/local/go
COPY --from=builder /usr/lib/jvm/default-java /usr/lib/jvm/default-java
COPY --from=builder /root/.cargo /root/.cargo
COPY --from=builder /usr/local/lib/python3.9/site-packages /usr/local/lib/python3.9/site-packages

RUN mkdir -p /AlgoRank/Solutions/{C,CPP,Java,Go,Rust}_Solutions \
    && mkdir -p /AlgoRank/Problem

WORKDIR /AlgoRank

COPY . .

ENTRYPOINT ["python3", "-u", "main.py"]