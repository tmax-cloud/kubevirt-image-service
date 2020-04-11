FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/kubevirt-image-service \
    USER_UID=1001 \
    USER_NAME=kubevirt-image-service

# install operator binary
COPY build/_output/bin/kubevirt-image-service ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
