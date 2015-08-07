FROM ubuntu:14.04
COPY vctl /vctl
COPY versionctl.html /versionctl.html
CMD ["./vctl"]
