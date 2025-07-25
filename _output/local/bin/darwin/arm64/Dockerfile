FROM ghcr.io/siderolabs/kubelet:v1.32.7
COPY --from=busybox:musl /bin/* /bin/
COPY kubelet /usr/local/bin/kubelet
