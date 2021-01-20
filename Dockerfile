FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/
EXPOSE 10000
ADD main /
CMD ["/main"]