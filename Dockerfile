FROM scratch
ADD ca-bundle.crt /etc/pki/tls/certs/
EXPOSE 10000
ADD main /
CMD ["/main"]