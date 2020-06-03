FROM registry.fedoraproject.org/fedora:31 as builder

WORKDIR /tmp/mantle
COPY . /tmp/mantle

RUN yum install -y --nodocs make golang gcc && \
    make && \
    make install DESTDIR=/tmp/mantle/ && \
    cp /tmp/mantle/default-run.sh /tmp/mantle/usr/bin/ && \
    chmod a+x /tmp/mantle/usr/bin/default-run.sh

FROM registry.fedoraproject.org/fedora-minimal:31
COPY --from=builder /tmp/mantle/usr/bin/* /usr/bin

CMD ["/usr/bin/default-run.sh"]
