FROM alpine:3.9

USER root
COPY . /
WORKDIR /

RUN apk add --no-cache python3 python3-dev libffi-dev openssl-dev curl build-base && \
    python3 -m ensurepip && \
    pip3 install -r requirements.txt && \
    rm -fr /root/.cache


ENTRYPOINT ["/usr/bin/python3", "app.py"]
